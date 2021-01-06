package controller

import (
	"github.com/akarasz/yahtzee/model"
)

type AddPlayerResponse struct {
	Players []*model.Player
}

func NewAddPlayerResponse(g *model.Game) *AddPlayerResponse {
	return &AddPlayerResponse{
		Players: g.Players,
	}
}

type RollResponse struct {
	Dices     []*model.Dice
	RollCount int
}

func NewRollResponse(g *model.Game) *RollResponse {
	return &RollResponse{
		Dices:     g.Dices,
		RollCount: g.RollCount,
	}
}

type LockResponse struct {
	Dices []*model.Dice
}

func NewLockResponse(g *model.Game) *LockResponse {
	return &LockResponse{
		Dices: g.Dices,
	}
}

type ScoreResponse struct {
	model.Game
}

func NewScoreResponse(g *model.Game) *ScoreResponse {
	return &ScoreResponse{*g}
}
