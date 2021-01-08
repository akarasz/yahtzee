package store

import (
	"errors"
	"sync"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee"
)

var (
	// ErrNotExists is returned when an ID not found in the store.
	ErrNotExists = errors.New("not exists")
)

// Store contains game elements by their IDs.
type Store interface {
	// Load returns a game from the store.
	Load(id string) (yahtzee.Game, error)

	// Save adds the game to the store.
	Save(id string, g yahtzee.Game) error
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

	empty := *yahtzee.NewGame()
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

func (ts *TestSuite) newAdvancedGame() *yahtzee.Game {
	return &yahtzee.Game{
		Players: []*yahtzee.Player{
			{
				User: yahtzee.User("Alice"),
				ScoreSheet: map[yahtzee.Category]int{
					yahtzee.Twos:      6,
					yahtzee.Fives:     15,
					yahtzee.FullHouse: 25,
				},
			}, {
				User: yahtzee.User("Bob"),
				ScoreSheet: map[yahtzee.Category]int{
					yahtzee.Threes:      6,
					yahtzee.FourOfAKind: 16,
				},
			}, {
				User: yahtzee.User("Carol"),
				ScoreSheet: map[yahtzee.Category]int{
					yahtzee.Twos:          6,
					yahtzee.SmallStraight: 30,
				},
			},
		},
		Dices: []*yahtzee.Dice{
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
