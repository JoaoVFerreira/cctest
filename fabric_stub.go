package cctest

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// MockStub extends Fabric's shimtest.MockStub with cctest query and event behavior.
type MockStub struct {
	*shimtest.MockStub
	args    [][]byte
	events  *EventLog
	history map[string][]*queryresult.KeyModification
}

var _ shim.ChaincodeStubInterface = (*MockStub)(nil)

func newMockStub(name string, cc shim.Chaincode, events *EventLog) *MockStub {
	return &MockStub{
		MockStub: shimtest.NewMockStub(name, cc),
		events:   events,
		history:  map[string][]*queryresult.KeyModification{},
	}
}

func (s *MockStub) setArgs(args [][]byte) {
	s.args = make([][]byte, len(args))
	for i, arg := range args {
		s.args[i] = cloneBytes(arg)
	}
}

// GetArgs returns the args set by Context.MockInvoke.
func (s *MockStub) GetArgs() [][]byte {
	out := make([][]byte, len(s.args))
	for i, arg := range s.args {
		out[i] = cloneBytes(arg)
	}
	return out
}

// GetStringArgs returns the args as strings.
func (s *MockStub) GetStringArgs() []string {
	args := s.GetArgs()
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, string(arg))
	}
	return out
}

// GetFunctionAndParameters splits the first arg as the function name.
func (s *MockStub) GetFunctionAndParameters() (string, []string) {
	args := s.GetStringArgs()
	if len(args) == 0 {
		return "", []string{}
	}
	return args[0], args[1:]
}

// GetArgsSlice joins raw args for compatibility with shim.ChaincodeStubInterface.
func (s *MockStub) GetArgsSlice() ([]byte, error) {
	var out []byte
	for _, arg := range s.args {
		out = append(out, arg...)
	}
	return out, nil
}

// SetEvent records chaincode events while preserving shimtest behavior.
func (s *MockStub) SetEvent(name string, payload []byte) error {
	if s.events != nil {
		s.events.record(name, payload)
	}
	return s.MockStub.SetEvent(name, payload)
}

// PutState writes state and records basic history.
func (s *MockStub) PutState(key string, value []byte) error {
	if err := s.MockStub.PutState(key, value); err != nil {
		return err
	}
	s.recordHistory(key, value, false)
	return nil
}

// DelState deletes state and records basic history.
func (s *MockStub) DelState(key string) error {
	if err := s.MockStub.DelState(key); err != nil {
		return err
	}
	s.recordHistory(key, nil, true)
	return nil
}

// DelPrivateData deletes private data from the in-memory collection.
func (s *MockStub) DelPrivateData(collection, key string) error {
	if _, ok := s.PvtState[collection]; !ok {
		return nil
	}
	delete(s.PvtState[collection], key)
	return nil
}

// GetQueryResult supports a conservative CouchDB selector subset.
func (s *MockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	kvs, err := s.queryState(query, s.State)
	if err != nil {
		return nil, err
	}
	return newStateIterator(kvs), nil
}

// GetQueryResultWithPagination supports a conservative CouchDB selector subset with page slicing.
func (s *MockStub) GetQueryResultWithPagination(query string, pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	kvs, err := s.queryStateWithPagination(query, s.State, pageSize, bookmark)
	if err != nil {
		return nil, nil, err
	}

	nextBookmark := ""
	if pageSize > 0 && len(kvs) == int(pageSize) {
		nextBookmark = kvs[len(kvs)-1].Key
	}
	return newStateIterator(kvs), &peer.QueryResponseMetadata{Bookmark: nextBookmark, FetchedRecordsCount: int32(len(kvs))}, nil
}

// GetPrivateDataByRange returns private data keys in lexical order.
func (s *MockStub) GetPrivateDataByRange(collection, startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	return newStateIterator(rangeKVs(s.PvtState[collection], startKey, endKey)), nil
}

// GetPrivateDataByPartialCompositeKey returns private composite keys matching the prefix.
func (s *MockStub) GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	prefix, err := s.CreateCompositeKey(objectType, keys)
	if err != nil {
		return nil, err
	}

	var out []*queryresult.KV
	for key, value := range s.PvtState[collection] {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			out = append(out, &queryresult.KV{Key: key, Value: cloneBytes(value)})
		}
	}
	sortKVs(out)
	return newStateIterator(out), nil
}

// GetPrivateDataQueryResult applies the selector subset to a private collection.
func (s *MockStub) GetPrivateDataQueryResult(collection, query string) (shim.StateQueryIteratorInterface, error) {
	kvs, err := s.queryState(query, s.PvtState[collection])
	if err != nil {
		return nil, err
	}
	return newStateIterator(kvs), nil
}

// GetHistoryForKey returns state updates recorded through this mock.
func (s *MockStub) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	items := make([]*queryresult.KeyModification, 0, len(s.history[key]))
	for _, item := range s.history[key] {
		items = append(items, cloneHistoryItem(item))
	}
	return &historyIterator{items: items}, nil
}

func (s *MockStub) clearState() {
	for key := range s.State {
		delete(s.State, key)
	}
	s.Keys.Init()
}

func (s *MockStub) queryState(query string, state map[string][]byte) ([]*queryresult.KV, error) {
	return s.queryStateWithPagination(query, state, 0, "")
}

func (s *MockStub) queryStateWithPagination(query string, state map[string][]byte, pageSize int32, bookmark string) ([]*queryresult.KV, error) {
	if state == nil {
		return nil, nil
	}

	spec, err := parseQuery(query)
	if err != nil {
		return nil, err
	}
	if bookmark != "" {
		spec.Bookmark = bookmark
	}
	if pageSize > 0 {
		spec.Limit = int(pageSize)
	}

	var out []*queryresult.KV
	for key, value := range state {
		var doc map[string]any
		if err := json.Unmarshal(value, &doc); err != nil {
			continue
		}
		matched, err := matchesSelector(doc, spec.Selector)
		if err != nil {
			return nil, err
		}
		if matched {
			out = append(out, &queryresult.KV{Key: key, Value: cloneBytes(value)})
		}
	}
	sortQueryResults(out, spec.Sort)
	if spec.Limit > 0 || spec.Bookmark != "" {
		page, _ := paginateKVs(out, int32(spec.Limit), spec.Bookmark)
		return page, nil
	}
	return out, nil
}

func parseQuery(query string) (querySpec, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(query), &raw); err != nil {
		return querySpec{}, fmt.Errorf("unsupported rich query: invalid JSON: %w", err)
	}
	selector, ok := raw["selector"].(map[string]any)
	if !ok || selector == nil {
		return querySpec{}, fmt.Errorf("unsupported rich query: missing selector")
	}

	spec := querySpec{Selector: selector}
	if rawLimit, ok := raw["limit"]; ok {
		limit, ok := intFromAny(rawLimit)
		if !ok {
			return querySpec{}, fmt.Errorf("unsupported rich query: limit must be a number")
		}
		spec.Limit = limit
	}
	if rawBookmark, ok := raw["bookmark"]; ok {
		bookmark, ok := rawBookmark.(string)
		if !ok {
			return querySpec{}, fmt.Errorf("unsupported rich query: bookmark must be a string")
		}
		spec.Bookmark = bookmark
	}
	if rawSort, ok := raw["sort"]; ok {
		sortFields, err := parseSort(rawSort)
		if err != nil {
			return querySpec{}, err
		}
		spec.Sort = sortFields
	}
	return spec, nil
}

type querySpec struct {
	Selector map[string]any
	Sort     []sortField
	Limit    int
	Bookmark string
}

type sortField struct {
	Field string
	Desc  bool
}

func parseSort(raw any) ([]sortField, error) {
	parts, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("unsupported rich query: sort must be an array")
	}
	fields := make([]sortField, 0, len(parts))
	for _, part := range parts {
		partMap, ok := part.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unsupported rich query: sort entries must be objects")
		}
		for field, directionAny := range partMap {
			direction, ok := directionAny.(string)
			if !ok {
				return nil, fmt.Errorf("unsupported rich query: sort direction for %s must be a string", field)
			}
			switch strings.ToLower(direction) {
			case "asc":
				fields = append(fields, sortField{Field: field})
			case "desc":
				fields = append(fields, sortField{Field: field, Desc: true})
			default:
				return nil, fmt.Errorf("unsupported rich query: sort direction %q", direction)
			}
		}
	}
	return fields, nil
}

func (s *MockStub) recordHistory(key string, value []byte, deleted bool) {
	ts := s.TxTimestamp
	if ts == nil {
		ts = ptypes.TimestampNow()
	}
	txID := s.TxID
	if txID == "" {
		txID = "cctest"
	}
	s.history[key] = append(s.history[key], &queryresult.KeyModification{
		TxId:      txID,
		Value:     cloneBytes(value),
		Timestamp: ts,
		IsDelete:  deleted,
	})
}

func cloneHistoryItem(item *queryresult.KeyModification) *queryresult.KeyModification {
	if item == nil {
		return nil
	}
	copy := *item
	copy.Value = cloneBytes(item.Value)
	return &copy
}

type historyIterator struct {
	items  []*queryresult.KeyModification
	cursor int
}

var _ shim.HistoryQueryIteratorInterface = (*historyIterator)(nil)

func (it *historyIterator) HasNext() bool {
	return it.cursor < len(it.items)
}

func (it *historyIterator) Close() error {
	return nil
}

func (it *historyIterator) Next() (*queryresult.KeyModification, error) {
	if !it.HasNext() {
		return nil, fmt.Errorf("history iterator exhausted")
	}
	item := it.items[it.cursor]
	it.cursor++
	return cloneHistoryItem(item), nil
}

func sortQueryResults(kvs []*queryresult.KV, fields []sortField) {
	if len(fields) == 0 {
		sortKVs(kvs)
		return
	}
	sort.SliceStable(kvs, func(i, j int) bool {
		left := decodeKV(kvs[i])
		right := decodeKV(kvs[j])
		for _, field := range fields {
			leftValues := valuesAtPath(left, field.Field)
			rightValues := valuesAtPath(right, field.Field)
			cmp := comparePathValues(leftValues, rightValues)
			if cmp == 0 {
				continue
			}
			if field.Desc {
				return cmp > 0
			}
			return cmp < 0
		}
		return kvs[i].Key < kvs[j].Key
	})
}

func decodeKV(kv *queryresult.KV) map[string]any {
	var out map[string]any
	_ = json.Unmarshal(kv.Value, &out)
	return out
}

func comparePathValues(left, right []any) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	if len(left) == 0 {
		return -1
	}
	if len(right) == 0 {
		return 1
	}
	cmp, ok := compareValues(left[0], right[0])
	if !ok {
		return 0
	}
	return cmp
}

func intFromAny(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	default:
		return 0, false
	}
}

type stateIterator struct {
	items  []*queryresult.KV
	cursor int
}

var _ shim.StateQueryIteratorInterface = (*stateIterator)(nil)

func newStateIterator(items []*queryresult.KV) *stateIterator {
	return &stateIterator{items: items}
}

func (it *stateIterator) HasNext() bool {
	return it.cursor < len(it.items)
}

func (it *stateIterator) Close() error {
	return nil
}

func (it *stateIterator) Next() (*queryresult.KV, error) {
	if !it.HasNext() {
		return nil, fmt.Errorf("iterator exhausted")
	}
	item := it.items[it.cursor]
	it.cursor++
	return item, nil
}

func rangeKVs(state map[string][]byte, startKey, endKey string) []*queryresult.KV {
	var out []*queryresult.KV
	for key, value := range state {
		if startKey != "" && key < startKey {
			continue
		}
		if endKey != "" && key >= endKey {
			continue
		}
		out = append(out, &queryresult.KV{Key: key, Value: cloneBytes(value)})
	}
	sortKVs(out)
	return out
}

func sortKVs(kvs []*queryresult.KV) {
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Key < kvs[j].Key
	})
}

func paginateKVs(kvs []*queryresult.KV, pageSize int32, bookmark string) ([]*queryresult.KV, string) {
	start := 0
	if bookmark != "" {
		for i, kv := range kvs {
			if kv.Key > bookmark {
				start = i
				break
			}
			start = i + 1
		}
	}
	if pageSize <= 0 {
		return kvs[start:], ""
	}

	end := start + int(pageSize)
	if end > len(kvs) {
		end = len(kvs)
	}

	next := ""
	if end < len(kvs) && end > start {
		next = kvs[end-1].Key
	}
	return kvs[start:end], next
}
