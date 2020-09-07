package controller

import (
	"github.com/akarasz/yahtzee/models"
)

type addPlayerResponse struct {
	Players []*models.Player
}

func newAddPlayerResponse(g *models.Game) *addPlayerResponse {
	return &addPlayerResponse{
		Players: g.Players,
	}
}

type rollResponse struct {
	Dices     []*models.Dice
	RollCount int
}

func newRollResponse(g *models.Game) *rollResponse {
	return &rollResponse{
		Dices:     g.Dices,
		RollCount: g.RollCount,
	}
}

type lockResponse struct {
	Dices []*models.Dice
}

func newLockResponse(g *models.Game) *lockResponse {
	return &lockResponse{
		Dices: g.Dices,
	}
}

type scoreResponse struct {
	models.Game
}

func newScoreResponse(g *models.Game) *scoreResponse {
	return &scoreResponse{*g}
}
