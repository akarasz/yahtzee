package controller

import (
	"math/rand"
	"strconv"

	"github.com/akarasz/yahtzee/models"
	"github.com/akarasz/yahtzee/service"
	"github.com/akarasz/yahtzee/store"
)

type Root interface {
	Create() (string, error)
	Get(gameID string) (*models.Game, error)
	Scores(dices []string) (map[models.Category]int, error)
}

type Game interface {
	AddPlayer(u *models.User, gameID string) (*addPlayerResponse, error)
	Roll(u *models.User, gameID string) (*rollResponse, error)
	Lock(u *models.User, gameID string, dice int) (*lockResponse, error)
	Score(u *models.User, gameID string, c models.Category) (*scoreResponse, error)
}

type Default struct {
	store           store.Store
	serviceProvider service.Provider
}

func (c *Default) Create() (string, error) {
	id := generateID()
	err := c.store.Save(id, *models.NewGame())
	return id, err
}

func (c *Default) Get(gameID string) (*models.Game, error) {
	g, err := c.store.Load(gameID)
	return &g, err
}

func (c *Default) Scores(dices []string) (map[models.Category]int, error) {
	dd := make([]int, 5)
	for i, d := range dices {
		v, err := strconv.Atoi(d)
		if err != nil {
			return nil, err
		}
		if err != nil || v < 1 || 6 < v {
			return nil, service.ErrInvalidDice
		}

		dd[i] = v
	}

	categories := []models.Category{
		models.Ones,
		models.Twos,
		models.Threes,
		models.Fours,
		models.Fives,
		models.Sixes,
		models.ThreeOfAKind,
		models.FourOfAKind,
		models.FullHouse,
		models.SmallStraight,
		models.LargeStraight,
		models.Yahtzee,
		models.Chance,
	}

	result := map[models.Category]int{}
	for _, c := range categories {
		score, err := service.Score(c, dd)
		if err != nil {
			return nil, err
		}
		result[c] = score
	}

	return result, nil
}

func (c *Default) AddPlayer(u *models.User, gameID string) (*addPlayerResponse, error) {
	g, err := c.store.Load(gameID)
	if err != nil {
		return nil, err
	}

	s := c.serviceProvider.Create(g, *u)
	res, err := s.AddPlayer()
	if err != nil {
		return nil, err
	}

	if err := c.store.Save(gameID, res); err != nil {
		return nil, err
	}

	return newAddPlayerResponse(&g), nil
}

func (c *Default) Roll(u *models.User, gameID string) (*rollResponse, error) {
	g, err := c.store.Load(gameID)
	if err != nil {
		return nil, err
	}

	s := c.serviceProvider.Create(g, *u)
	res, err := s.Roll()
	if err != nil {
		return nil, err
	}

	if err := c.store.Save(gameID, res); err != nil {
		return nil, err
	}

	return newRollResponse(&g), nil
}

func (c *Default) Lock(u *models.User, gameID string, dice int) (*lockResponse, error) {
	g, err := c.store.Load(gameID)
	if err != nil {
		return nil, err
	}

	s := c.serviceProvider.Create(g, *u)
	res, err := s.Lock(dice)
	if err != nil {
		return nil, err
	}

	if err := c.store.Save(gameID, res); err != nil {
		return nil, err
	}

	return newLockResponse(&g), nil
}

func (c *Default) Score(u *models.User, gameID string, category models.Category) (*scoreResponse, error) {
	g, err := c.store.Load(gameID)
	if err != nil {
		return nil, err
	}

	s := c.serviceProvider.Create(g, *u)
	res, err := s.Score(category)
	if err != nil {
		return nil, err
	}

	if err := c.store.Save(gameID, res); err != nil {
		return nil, err
	}

	return newScoreResponse(&g), nil
}

func generateID() string {
	const (
		idCharset = "abcdefghijklmnopqrstvwxyz0123456789"
		length    = 4
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}
