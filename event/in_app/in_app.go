package in_app

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/model"
)

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

func (b *InApp) Emit(gameID string, u *model.User, t event.Type, body interface{}) {
	g, ok := b.games[gameID]
	if !ok {
		return
	}

	g.Lock()
	defer g.Unlock()

	for _, s := range g.clients {
		s <- event.Event{
			User:   u,
			Action: t,
			Data:   body,
		}
	}
}
