package store

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee/models"
)

var (
	// ErrNotExists is returned when an ID not found in the store.
	ErrNotExists = errors.New("not exists")
)

// Store contains game elements by their IDs.
type Store interface {
	// Load returns a game from the store.
	Load(id string) (models.Game, error)

	// Save adds the game to the store.
	Save(id string, g models.Game) error
}

// InMemory is the in-memory implementation of Store.
type InMemory struct {
	repo map[string]models.Game
}

func (s *InMemory) Save(id string, g models.Game) error {
	s.repo[id] = g

	return nil
}

func (s *InMemory) Load(id string) (models.Game, error) {
	g, ok := s.repo[id]
	if !ok {
		return g, ErrNotExists
	}

	return g, nil
}

// NewInMemory creates an empty in-memory store.
func New() *InMemory {
	res := InMemory{
		repo: map[string]models.Game{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_store_size",
			Help: "The total number of games in the in memory store",
		},
		func() float64 { return float64(len(res.repo)) })

	return &res
}
