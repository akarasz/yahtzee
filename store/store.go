package store

import (
	"errors"

	"github.com/akarasz/yahtzee/models"
)

var (
	// ErrAlreadyExists is returned when an entry is already in the store with the given ID.
	ErrAlreadyExists = errors.New("already exists")

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
	if _, ok := s.repo[id]; ok {
		return ErrAlreadyExists
	}

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
func NewInMemory() *InMemory {
	return &InMemory{
		repo: map[string]models.Game{},
	}
}
