package handler

import (
	// "net/http"
	// "net/http/httptest"
	// "testing"

	"github.com/akarasz/yahtzee/pkg/models"
)

type controllerStub struct {
	snapshotReturns *models.Game
}

func (s *controllerStub) AddPlayer(name string) error {
	return nil
}

func (s *controllerStub) Roll(player string) ([]*models.Dice, error) {
	return nil, nil
}

func (s *controllerStub) Toggle(player string, diceIndex int) ([]*models.Dice, error) {
	return nil, nil
}

func (s *controllerStub) Score(player string, c models.Category) error {
	return nil
}

func (s *controllerStub) Snapshot() *models.Game {
	return s.snapshotReturns
}
