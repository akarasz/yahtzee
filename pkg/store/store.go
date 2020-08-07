package store

import (
	"errors"

	"github.com/akarasz/yahtzee/pkg/game"
)

var (
	// ErrAlreadyExists is returned when an entry is already in the store with the given ID.
	ErrAlreadyExists = errors.New("already exists")

	// ErrNotExists is returned when an ID not found in the store.
	ErrNotExists = errors.New("not exists")
)

// Store contains game elements by their IDs.
type Store interface {
	Get(id string) (*game.Game, error)
	Put(id string, g *game.Game) error
}

// InMemory is the in-memory implementation of Store.
type InMemory struct {
	repo map[string]*game.Game
}

// Put adds the game to the store.
func (s *InMemory) Put(id string, g *game.Game) error {
	if _, ok := s.repo[id]; ok {
		return ErrAlreadyExists
	}

	s.repo[id] = g

	return nil
}

// Get returns a game from the store.
func (s *InMemory) Get(id string) (*game.Game, error) {
	g, ok := s.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	return g, nil
}

// NewInMemory creates an empty in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{
		repo: map[string]*game.Game{},
	}
}
