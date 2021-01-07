package event

import (
	"fmt"
	"sync"
	"time"

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
	s := ts.S
	e := ts.E

	c, err := s.Subscribe("subscribeID", "subscribeWSID")
	ts.NoError(err)

	got := ts.receiveWithTimeout(c)
	e.Emit("subscribeID", model.NewUser("Alice"), AddPlayer, nil)
	ts.NotNil(<-got)
}

func (ts *TestSuite) TestUnsubscribe() {
	s := ts.S
	e := ts.E

	c, err := s.Subscribe("unsubscribeID", "unsubscribeWSID")
	ts.Require().NoError(err)

	ts.NoError(s.Unsubscribe("unsubscribeID", "unsubscribeWSID"))

	got := ts.receiveWithTimeout(c)
	e.Emit("unsubscribeID", model.NewUser("Alice"), AddPlayer, nil)
	ts.Nil(<-got)
}

func (ts *TestSuite) TestEmit() {
	s := ts.S
	e := ts.E

	c1, err := s.Subscribe("emitID", "emit1WSID")
	ts.Require().NoError(err)
	c2, err := s.Subscribe("emitID", "emit2WSID")
	ts.Require().NoError(err)
	c3, err := s.Subscribe("notEmitID", "emit3WSID")
	ts.Require().NoError(err)

	got1 := ts.receiveWithTimeout(c1)
	got2 := ts.receiveWithTimeout(c2)
	got3 := ts.receiveWithTimeout(c3)
	e.Emit("emitID", model.NewUser("Alice"), AddPlayer, nil)
	ts.NotNil(<-got1)
	ts.NotNil(<-got2)
	ts.Nil(<-got3)
}

func (ts *TestSuite) TestRace() {
	s := ts.S
	e := ts.E
	wg := &sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			id := fmt.Sprintf("raceID%d", i)
			c, err := s.Subscribe(id, id+"WS")
			ts.Require().NoError(err)

			go func(c chan interface{}) {
				for {
					<-c
				}
			}(c)

			for j := 0; j < 3; j++ {
				e.Emit(id, model.NewUser("Alice"), AddPlayer, nil)
			}

			ts.Require().NoError(s.Unsubscribe(id, id+"WS"))
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (ts *TestSuite) receiveWithTimeout(c <-chan interface{}) chan interface{} {
	res := make(chan interface{}, 1)

	go func() {
		for {
			select {
			case got := <-c:
				res <- got
			case <-time.After(100 * time.Millisecond):
				res <- nil
				return
			}
		}
	}()

	return res
}
