package events

//go:generate mockgen -destination=mocks/mock_events.go -package=mocks -build_flags=-mod=mod . Emitter,Subscriber

import (
	"errors"
	"sync"
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
	// Emit notifies the consumers of `gameID` that a `t` event happened
	// with the changes described in `body`
	Emit(gameID string, t Type, body interface{})
}

func New() *Broker {
	return &Broker{
		clients: map[string]*game{},
	}
}

type game struct {
	sync.Mutex
	clients map[interface{}]chan interface{}
}

func newGame() *game {
	return &game{
		clients: map[interface{}]chan interface{}{},
	}
}

type Broker struct {
	sync.Mutex
	clients map[string]*game
}

func (b *Broker) Subscribe(gameID string, clientID interface{}) (chan interface{}, error) {
	c := make(chan interface{})

	var g *game

	g, ok := b.clients[gameID]
	if !ok {
		b.Lock()
		defer b.Unlock()

		g = newGame()
		b.clients[gameID] = g
	}

	g.Lock()
	defer g.Unlock()

	g.clients[clientID] = c

	return c, nil
}

func (b *Broker) Unsubscribe(gameID string, clientID interface{}) error {
	g, ok := b.clients[gameID]
	if !ok {
		return errors.New("no game found")
	}

	g.Lock()
	defer g.Unlock()

	if c, ok := g.clients[clientID]; ok {
		close(c)
		delete(g.clients, clientID)
	}

	return nil
}

func (b *Broker) Emit(gameID string, t Type, body interface{}) {
	g, ok := b.clients[gameID]
	if !ok {
		return
	}

	g.Lock()
	defer g.Unlock()

	for _, s := range g.clients {
		s <- body
	}
}
