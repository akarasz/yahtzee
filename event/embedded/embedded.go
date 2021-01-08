package embedded

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/event"
)

type game struct {
	sync.Mutex
	clients map[interface{}]chan *event.Event
}

func newGame() *game {
	return &game{
		clients: map[interface{}]chan *event.Event{},
	}
}

type InApp struct {
	sync.RWMutex
	games map[string]*game
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
		func() float64 {
			res.RLock()
			total := len(res.games)
			res.RUnlock()
			return float64(total)
		})

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_websocket_clients_total",
			Help: "The total number of clients with websocket channels",
		},
		func() float64 {
			total := 0
			res.RLock()
			for _, g := range res.games {
				total += len(g.clients)
			}
			res.RUnlock()
			return float64(total)
		})

	return &res
}

func (b *InApp) Subscribe(gameID string, clientID interface{}) (chan *event.Event, error) {
	b.Lock()
	defer b.Unlock()

	c := make(chan *event.Event)

	var g *game

	g, ok := b.games[gameID]
	if !ok {

		g = newGame()
		b.games[gameID] = g
	}

	g.Lock()
	g.clients[clientID] = c
	g.Unlock()

	return c, nil
}

func (b *InApp) Unsubscribe(gameID string, clientID interface{}) error {
	b.Lock()
	defer b.Unlock()

	g, ok := b.games[gameID]
	if !ok {
		return errors.New("no game found")
	}

	g.Lock()
	if c, ok := g.clients[clientID]; ok {
		close(c)
		delete(g.clients, clientID)
	}

	if len(g.clients) == 0 {
		delete(b.games, gameID)
	}
	g.Unlock()

	return nil
}

func (b *InApp) Emit(gameID string, u *yahtzee.User, t event.Type, body interface{}) {
	b.RLock()
	g, ok := b.games[gameID]
	b.RUnlock()
	if !ok {
		return
	}

	g.Lock()
	defer g.Unlock()

	for _, s := range g.clients {
		s <- &event.Event{
			User:   u,
			Action: t,
			Data:   body,
		}
	}
}
