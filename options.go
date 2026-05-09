package cctest

import "github.com/hyperledger/fabric-chaincode-go/shim"

// Option configures every Context created by a suite.
type Option func(*Config)

// Config is the mutable suite configuration used by options.
type Config struct {
	chaincode        shim.Chaincode
	mspID            string
	channelID        string
	identity         ClientIdentity
	initialState     map[string][]byte
	initialJSONState map[string]any
	initArgs         [][]byte
	contextSetups    []func(*Context) error
	prettyOutput     bool
	colorOutput      *bool
}

func defaultSuiteConfig() Config {
	return Config{
		mspID:     "Org1MSP",
		channelID: "cctest",
		identity: ClientIdentity{
			MSPID:      "Org1MSP",
			ID:         "cctest-user",
			Attributes: map[string]string{},
		},
		prettyOutput: true,
	}
}

// SetChaincode configures the chaincode invoked by Context.MockInvoke.
func (c *Config) SetChaincode(cc shim.Chaincode) {
	c.chaincode = cc
}

// AddContextSetup registers a setup function that runs for every Context.
func (c *Config) AddContextSetup(fn func(*Context) error) {
	if fn != nil {
		c.contextSetups = append(c.contextSetups, fn)
	}
}

// WithChaincode configures the chaincode invoked by Context.MockInvoke.
func WithChaincode(cc shim.Chaincode) Option {
	return func(config *Config) {
		config.SetChaincode(cc)
	}
}

// WithMSP configures the default MSP ID.
func WithMSP(mspID string) Option {
	return func(config *Config) {
		config.mspID = mspID
		if config.identity.MSPID == "" || config.identity.MSPID == "Org1MSP" {
			config.identity.MSPID = mspID
		}
	}
}

// WithChannelID configures the mock channel ID.
func WithChannelID(channelID string) Option {
	return func(config *Config) {
		config.channelID = channelID
	}
}

// WithClientIdentity configures the client identity available to chaincode.
func WithClientIdentity(identity ClientIdentity) Option {
	return func(config *Config) {
		config.identity = identity
		if config.identity.MSPID == "" {
			config.identity.MSPID = config.mspID
		}
		if config.identity.ID == "" {
			config.identity.ID = "cctest-user"
		}
		if config.identity.Attributes == nil {
			config.identity.Attributes = map[string]string{}
		}
	}
}

// WithInitialState seeds public ledger state for each It.
func WithInitialState(state map[string][]byte) Option {
	return func(config *Config) {
		config.initialState = cloneBytesMap(state)
	}
}

// WithInitialJSONState marshals and seeds public JSON ledger state for each It.
func WithInitialJSONState(state map[string]any) Option {
	return func(config *Config) {
		config.initialJSONState = cloneAnyMap(state)
	}
}

// WithInitArgs runs the configured chaincode Init for each Context.
func WithInitArgs(args ...[]byte) Option {
	return func(config *Config) {
		config.initArgs = cloneArgs(args)
	}
}

// WithPrettyOutput controls cctest's Jest-like verbose reporter.
func WithPrettyOutput(enabled bool) Option {
	return func(config *Config) {
		config.prettyOutput = enabled
	}
}

// WithColorOutput controls ANSI colors in cctest's verbose reporter.
func WithColorOutput(enabled bool) Option {
	return func(config *Config) {
		config.colorOutput = &enabled
	}
}

func cloneBytesMap(input map[string][]byte) map[string][]byte {
	if input == nil {
		return nil
	}
	out := make(map[string][]byte, len(input))
	for key, value := range input {
		out[key] = cloneBytes(value)
	}
	return out
}

func cloneAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	out := make([]byte, len(value))
	copy(out, value)
	return out
}

func cloneArgs(args [][]byte) [][]byte {
	if args == nil {
		return nil
	}
	out := make([][]byte, len(args))
	for i, arg := range args {
		out[i] = cloneBytes(arg)
	}
	return out
}
