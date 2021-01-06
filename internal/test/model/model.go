package model

import (
	"github.com/akarasz/yahtzee/models"
)

func NewAdvanced() *models.Game {
	return &models.Game{
		Players: []*models.Player{
			{
				User: models.User("Alice"),
				ScoreSheet: map[models.Category]int{
					models.Twos:      6,
					models.Fives:     15,
					models.FullHouse: 25,
				},
			}, {
				User: models.User("Bob"),
				ScoreSheet: map[models.Category]int{
					models.Threes:      6,
					models.FourOfAKind: 16,
				},
			}, {
				User: models.User("Carol"),
				ScoreSheet: map[models.Category]int{
					models.Twos:          6,
					models.SmallStraight: 30,
				},
			},
		},
		Dices: []*models.Dice{
			{Value: 3, Locked: true},
			{Value: 2, Locked: false},
			{Value: 3, Locked: true},
			{Value: 1, Locked: false},
			{Value: 5, Locked: false},
		},
		Round:         5,
		CurrentPlayer: 1,
		RollCount:     1,
	}
}
