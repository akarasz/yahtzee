package event

import (
	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee/model"
)

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
	Subscribe(gameID string, clientID interface{}) (chan interface{}, error)
	Unsubscribe(gameID string, clientID interface{}) error
}

// Emitter used by the event producer side to fire events
type Emitter interface {
	// Emit notifies the consumers of `gameID` that `u` user triggered `t` event
	// that caused changes described in `body`
	Emit(gameID string, u *model.User, t Type, body interface{})
}

type Event struct {
	User   *model.User
	Action Type
	Data   interface{}
}

type TestSuite struct {
	suite.Suite

	S Subscriber
	E Emitter
}

func (ts *TestSuite) TestSubscribe() {
}

func (ts *TestSuite) TestUnsubscribe() {
}

func (ts *TestSuite) TestEmit() {
}
