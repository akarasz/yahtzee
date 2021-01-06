package store

import (
	"errors"

	"github.com/akarasz/yahtzee/model"
)

var (
	// ErrNotExists is returned when an ID not found in the store.
	ErrNotExists = errors.New("not exists")
)

// Store contains game elements by their IDs.
type Store interface {
	// Load returns a game from the store.
	Load(id string) (model.Game, error)

	// Save adds the game to the store.
	Save(id string, g model.Game) error
}