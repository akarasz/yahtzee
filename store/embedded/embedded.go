package embedded

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/store"
)

// InMemory is the in-memory implementation of Store.
type InMemory struct {
	repo map[string]yahtzee.Game
}

func (s *InMemory) Save(id string, g yahtzee.Game) error {
	s.repo[id] = g

	return nil
}

func (s *InMemory) Load(id string) (yahtzee.Game, error) {
	g, ok := s.repo[id]
	if !ok {
		return g, store.ErrNotExists
	}

	return g, nil
}

// NewInMemory creates an empty in-memory store.
func New() *InMemory {
	res := InMemory{
		repo: map[string]yahtzee.Game{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_store_size",
			Help: "The total number of games in the in memory store",
		},
		func() float64 { return float64(len(res.repo)) })

	return &res
}
