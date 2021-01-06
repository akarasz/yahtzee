package test

import (
	"github.com/stretchr/testify/suite"

	model_test "github.com/akarasz/yahtzee/internal/test/model"
	"github.com/akarasz/yahtzee/model"
	"github.com/akarasz/yahtzee/store"
)

type StoreTestSuite struct {
	suite.Suite

	subject store.Store
}

func New(s store.Store) *StoreTestSuite {
	return &StoreTestSuite{
		subject: s,
	}
}

func (ts *StoreTestSuite) TestLoad() {
	s := ts.subject

	_, err := s.Load("aaaaa")
	ts.Exactly(store.ErrNotExists, err)

	saved := *model_test.NewAdvanced()

	ts.Require().NoError(s.Save("aaaaa", saved))

	if got, err := s.Load("aaaaa"); ts.NoError(err) {
		ts.Exactly(saved, got)
	}
}

func (ts *StoreTestSuite) TestSave() {
	s := ts.subject

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
