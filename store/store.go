package store

import (
	"errors"
	"sync"

	"github.com/stretchr/testify/suite"

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

type TestSuite struct {
	suite.Suite

	Subject Store
}

func (ts *TestSuite) TestLoad() {
	s := ts.Subject

	_, err := s.Load("aaaaa")
	ts.Exactly(ErrNotExists, err)

	saved := *ts.newAdvancedGame()

	ts.Require().NoError(s.Save("aaaaa", saved))

	if got, err := s.Load("aaaaa"); ts.NoError(err) {
		ts.Exactly(saved, got)
	}
}

func (ts *TestSuite) TestSave() {
	s := ts.Subject

	empty := *model.NewGame()
	ts.NoError(s.Save("bbbbb", empty))

	if got, err := s.Load("bbbbb"); ts.NoError(err) {
		ts.Exactly(empty, got)
	}

	advanced := *ts.newAdvancedGame()
	ts.NoError(s.Save("bbbbb", advanced))

	if got, err := s.Load("bbbbb"); ts.NoError(err) {
		ts.Exactly(advanced, got)
	}
}

func (ts *TestSuite) TestRace() {
	s := ts.Subject
	wg := &sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			s.Save("ccccc", *ts.newAdvancedGame())
			s.Load("ccccc")
			wg.Done()
		}()
	}
	wg.Wait()
}

func (ts *TestSuite) newAdvancedGame() *model.Game {
	return &model.Game{
		Players: []*model.Player{
			{
				User: model.User("Alice"),
				ScoreSheet: map[model.Category]int{
					model.Twos:      6,
					model.Fives:     15,
					model.FullHouse: 25,
				},
			}, {
				User: model.User("Bob"),
				ScoreSheet: map[model.Category]int{
					model.Threes:      6,
					model.FourOfAKind: 16,
				},
			}, {
				User: model.User("Carol"),
				ScoreSheet: map[model.Category]int{
					model.Twos:          6,
					model.SmallStraight: 30,
				},
			},
		},
		Dices: []*model.Dice{
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
