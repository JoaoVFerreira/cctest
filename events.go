package cctest

import "testing"

// Event records a chaincode event emitted in a Context.
type Event struct {
	Name    string
	Payload []byte
}

// EventLog stores chaincode events for readable assertions.
type EventLog struct {
	t      *testing.T
	events []Event
}

func newEventLog(t *testing.T) *EventLog {
	return &EventLog{t: t}
}

func (l *EventLog) record(name string, payload []byte) {
	l.events = append(l.events, Event{Name: name, Payload: cloneBytes(payload)})
}

// All returns all recorded events.
func (l *EventLog) All() []Event {
	out := make([]Event, len(l.events))
	for i, event := range l.events {
		out[i] = Event{Name: event.Name, Payload: cloneBytes(event.Payload)}
	}
	return out
}

// Names returns event names in emission order.
func (l *EventLog) Names() []string {
	out := make([]string, 0, len(l.events))
	for _, event := range l.events {
		out = append(out, event.Name)
	}
	return out
}

// Last returns the last event and whether one exists.
func (l *EventLog) Last() (Event, bool) {
	if len(l.events) == 0 {
		return Event{}, false
	}
	event := l.events[len(l.events)-1]
	return Event{Name: event.Name, Payload: cloneBytes(event.Payload)}, true
}

// ExpectEmitted fails the test if name was not emitted.
func (l *EventLog) ExpectEmitted(name string) {
	l.t.Helper()
	for _, event := range l.events {
		if event.Name == name {
			return
		}
	}
	l.t.Fatalf("expected event %q to be emitted, got %v", name, l.Names())
}

// ExpectPayload fails the test if name was not emitted with payload.
func (l *EventLog) ExpectPayload(name string, payload []byte) {
	l.t.Helper()
	for _, event := range l.events {
		if event.Name == name && string(event.Payload) == string(payload) {
			return
		}
	}
	l.t.Fatalf("expected event %q payload %q, got %v", name, payload, l.events)
}
