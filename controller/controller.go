package controller

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/bsm/redislock"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/model"
	"github.com/akarasz/yahtzee/service"
	"github.com/akarasz/yahtzee/store"
)

var (
	lockExpiration = 5 * time.Second
	lockBackoff    = redislock.LinearBackoff(50 * time.Millisecond)
)

type Root interface {
	Create() (string, error)
	Get(gameID string) (*model.Game, error)
	Scores(dices []string) (map[model.Category]int, error)
}

type Game interface {
	AddPlayer(u *model.User, gameID string) (*AddPlayerResponse, error)
	Roll(u *model.User, gameID string) (*RollResponse, error)
	Lock(u *model.User, gameID string, dice string) (*LockResponse, error)
	Score(u *model.User, gameID string, c model.Category) (*ScoreResponse, error)
}

type Default struct {
	store           store.Store
	serviceProvider service.Provider
	event           event.Emitter
	locker          *redislock.Client
}

func New(s store.Store, p service.Provider, e event.Emitter, l *redislock.Client) *Default {
	return &Default{
		store:           s,
		serviceProvider: p,
		event:           e,
		locker:          l,
	}
}

func (c *Default) Create() (string, error) {
	id := generateID()
	err := c.store.Save(id, *model.NewGame())
	return id, err
}

func (c *Default) Get(gameID string) (*model.Game, error) {
	g, err := c.store.Load(gameID)
	return &g, err
}

func (c *Default) Scores(dices []string) (map[model.Category]int, error) {
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

	categories := []model.Category{
		model.Ones,
		model.Twos,
		model.Threes,
		model.Fours,
		model.Fives,
		model.Sixes,
		model.ThreeOfAKind,
		model.FourOfAKind,
		model.FullHouse,
		model.SmallStraight,
		model.LargeStraight,
		model.Yahtzee,
		model.Chance,
	}

	result := map[model.Category]int{}
	for _, c := range categories {
		score, err := service.Score(c, dd)
		if err != nil {
			return nil, err
		}
		result[c] = score
	}

	return result, nil
}

func (c *Default) lockGame(gameID string) (*redislock.Lock, error) {
	lock, err := c.locker.Obtain("lock:"+gameID, lockExpiration, &redislock.Options{
		RetryStrategy: lockBackoff,
	})

	if err != nil {
		if err == redislock.ErrNotObtained {
			log.Println("could not obtain lock")
		} else if err != nil {
			log.Fatalln(err)
		}
		return nil, err
	}

	return lock, nil
}

func (c *Default) AddPlayer(u *model.User, gameID string) (*AddPlayerResponse, error) {
	lock, err := c.lockGame(gameID)
	if err != nil {
		return nil, err
	}
	defer lock.Release()

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

	changes := NewAddPlayerResponse(&res)
	c.event.Emit(gameID, u, event.AddPlayer, changes)
	return changes, nil
}

func (c *Default) Roll(u *model.User, gameID string) (*RollResponse, error) {
	lock, err := c.lockGame(gameID)
	if err != nil {
		return nil, err
	}
	defer lock.Release()

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

	changes := NewRollResponse(&res)
	c.event.Emit(gameID, u, event.Roll, changes)
	return changes, nil
}

func (c *Default) Lock(u *model.User, gameID string, dice string) (*LockResponse, error) {
	lock, err := c.lockGame(gameID)
	if err != nil {
		return nil, err
	}
	defer lock.Release()

	diceIndex, err := strconv.Atoi(dice)
	if err != nil || diceIndex < 0 || diceIndex > 4 {
		return nil, service.ErrInvalidDice
	}

	g, err := c.store.Load(gameID)
	if err != nil {
		return nil, err
	}

	s := c.serviceProvider.Create(g, *u)
	res, err := s.Lock(diceIndex)
	if err != nil {
		return nil, err
	}

	if err := c.store.Save(gameID, res); err != nil {
		return nil, err
	}

	changes := NewLockResponse(&res)
	c.event.Emit(gameID, u, event.Lock, changes)
	return changes, nil
}

func (c *Default) Score(u *model.User, gameID string, category model.Category) (*ScoreResponse, error) {
	lock, err := c.lockGame(gameID)
	if err != nil {
		return nil, err
	}
	defer lock.Release()

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

	changes := NewScoreResponse(&res)
	c.event.Emit(gameID, u, event.Score, changes)
	return changes, nil
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
