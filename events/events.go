package events

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee/models"
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
	Emit(gameID string, u *models.User, t Type, body interface{})
}

type Event struct {
	User   *models.User
	Action Type
	Data   interface{}
}

func New() *InApp {
	res := InApp{
		games: map[string]*game{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_websocket_games_total",
			Help: "The total number of games with websocket channels",
		},
		func() float64 { return float64(len(res.games)) })

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_websocket_clients_total",
			Help: "The total number of clients with websocket channels",
		},
		func() float64 {
			total := 0
			for _, g := range res.games {
				total += len(g.clients)
			}
			return float64(total)
		})

	return &res
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

type InApp struct {
	sync.Mutex
	games map[string]*game
}

func (b *InApp) Subscribe(gameID string, clientID interface{}) (chan interface{}, error) {
	c := make(chan interface{})

	var g *game

	g, ok := b.games[gameID]
	if !ok {
		b.Lock()
		defer b.Unlock()

		g = newGame()
		b.games[gameID] = g
	}

	g.Lock()
	defer g.Unlock()

	g.clients[clientID] = c

	return c, nil
}

func (b *InApp) Unsubscribe(gameID string, clientID interface{}) error {
	g, ok := b.games[gameID]
	if !ok {
		return errors.New("no game found")
	}

	g.Lock()
	defer g.Unlock()

	if c, ok := g.clients[clientID]; ok {
		close(c)
		delete(g.clients, clientID)
	}

	if len(g.clients) == 0 {
		delete(b.games, gameID)
	}

	return nil
}

func (b *InApp) Emit(gameID string, u *models.User, t Type, body interface{}) {
	g, ok := b.games[gameID]
	if !ok {
		return
	}

	g.Lock()
	defer g.Unlock()

	for _, s := range g.clients {
		s <- Event{u, t, body}
	}
}
