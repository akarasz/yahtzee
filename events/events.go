package events

// Type tells which kind of events happened
type Type int

// Available types
const (
	AddPlayer Type = iota
	Roll
	Lock
	Score
)

// Emitter used by the event producer side to fire events
type Emitter interface {
	// Emit notifies the consumers of `gameID` that a `t` event happened
	// with the changes described in `body`
	Emit(gameID string, t Type, body interface{})
}
