package store

import (
	"errors"
	"sync"

	"github.com/stretchr/testify/suite"

	model_test "github.com/akarasz/yahtzee/internal/test/model"
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

	saved := *model_test.NewAdvanced()

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

	advanced := *model_test.NewAdvanced()
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
			s.Save("ccccc", *model_test.NewAdvanced())
			s.Load("ccccc")
			wg.Done()
		}()
	}
	wg.Wait()
}
