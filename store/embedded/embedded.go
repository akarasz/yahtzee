package embedded

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/store"
)

// InMemory is the in-memory implementation of Store.
type InMemory struct {
	repo map[string]yahtzee.Game
	locks map[string]*sync.Mutex

	repoLock *sync.RWMutex
	locksLock *sync.Mutex
}

func (s *InMemory) Save(id string, g yahtzee.Game) error {
	s.repoLock.Lock()
	s.repo[id] = g
	s.repoLock.Unlock()

	return nil
}

func (s *InMemory) Load(id string) (yahtzee.Game, error) {
	s.repoLock.RLock()
	g, ok := s.repo[id]
	s.repoLock.RUnlock()
	if !ok {
		return g, store.ErrNotExists
	}

	return g, nil
}

func (s *InMemory) Lock(id string) (func(), error) {
	s.locksLock.Lock()
	l, ok := s.locks[id]
	if !ok {
		l = &sync.Mutex{}
		s.locks[id] = l
	}
	s.locksLock.Unlock()

	l.Lock()

	return func() {
		l.Unlock()
	}, nil
}

// NewInMemory creates an empty in-memory store.
func New() *InMemory {
	res := InMemory{
		repo: map[string]yahtzee.Game{},
		locks: map[string]*sync.Mutex{},

		repoLock: &sync.RWMutex{},
		locksLock: &sync.Mutex{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_store_size",
			Help: "The total number of games in the in memory store",
		},
		func() float64 { return float64(len(res.repo)) })

	return &res
}
