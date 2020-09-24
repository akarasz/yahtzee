package events

//go:generate mockgen -destination=mocks/mock_events.go -package=mocks -build_flags=-mod=mod . Emitter,Subscriber

// Type tells which kind of events happened
type Type string

// Available types
const (
	AddPlayer Type = "add-player"
	Roll      Type = "roll"
	Lock      Type = "lock"
	Score     Type = "score"
)

// Subscriber for subscribe events
type Subscriber interface {
	// Subscribe to get events from `gameID` to be send to `channel`
	Subscribe(gameID string) (chan interface{}, error)
}

// Emitter used by the event producer side to fire events
type Emitter interface {
	// Emit notifies the consumers of `gameID` that a `t` event happened
	// with the changes described in `body`
	Emit(gameID string, t Type, body interface{})
}

type Dummy struct {
	C chan interface{}
}

func (e *Dummy) Subscribe(gameID string) (chan interface{}, error) {
	return e.C, nil
}

func (e *Dummy) Emit(gameID string, t Type, body interface{}) {
	e.C <- body
}
