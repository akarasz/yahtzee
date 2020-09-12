package controller

import (
	"github.com/akarasz/yahtzee/models"
)

type AddPlayerResponse struct {
	Players []*models.Player
}

func NewAddPlayerResponse(g *models.Game) *AddPlayerResponse {
	return &AddPlayerResponse{
		Players: g.Players,
	}
}

type RollResponse struct {
	Dices     []*models.Dice
	RollCount int
}

func NewRollResponse(g *models.Game) *RollResponse {
	return &RollResponse{
		Dices:     g.Dices,
		RollCount: g.RollCount,
	}
}

type LockResponse struct {
	Dices []*models.Dice
}

func NewLockResponse(g *models.Game) *LockResponse {
	return &LockResponse{
		Dices: g.Dices,
	}
}

type ScoreResponse struct {
	models.Game
}

func NewScoreResponse(g *models.Game) *ScoreResponse {
	return &ScoreResponse{*g}
}
