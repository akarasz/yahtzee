package yahtzee_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee"
	event "github.com/akarasz/yahtzee/event/embedded"
	"github.com/akarasz/yahtzee/model"
	store "github.com/akarasz/yahtzee/store/embedded"
)

type testSuite struct {
	suite.Suite

	store *store.InMemory
	event *event.InApp

	handler http.Handler
}

func TestSuite(t *testing.T) {
	s := store.New()
	e := event.New()

	suite.Run(t, &testSuite{
		store:   s,
		event:   e,
		handler: yahtzee.NewHandler(s, e, e),
	})
}

func (ts *testSuite) TestCreate() {
	rr := ts.newRequest("POST", "/")
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		got, err := ts.store.Load(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		ts.Require().NoError(err)
		ts.Exactly(*model.NewGame(), got)
	}
}

func (ts *testSuite) TestHints() {
	badInputs := []struct {
		description string
		query       string
	}{
		{"no query", "noop=true"},
		{"empty dices", "dices=1,2,3,4"},
		{"too few dices", "dices=1,2,3,4"},
		{"too many dices", "dices=1,2,3,4,5,6"},
		{"has low face value", "dices=1,1,1,0,1"},
		{"has high face value", "dices=7,6,6,6,6"},
	}
	for _, tc := range badInputs {
		rr := ts.newRequest("GET", "/score", tc.query)
		ts.Exactly(http.StatusBadRequest, rr.Code, "when %s", tc.description)
	}

	rr := ts.newRequest("GET", "/score", "dices=3,2,6,4,5")
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":0,
			"twos":2,
			"threes":3,
			"fours":4,
			"fives":5,
			"sixes":6,
			"three-of-a-kind":0,
			"four-of-a-kind":0,
			"full-house":0,
			"small-straight":30,
			"large-straight":40,
			"yahtzee":0,
			"chance":20
		}`, rr.Body.String())
}

func (ts *testSuite) TestGet() {
	rr := ts.newRequest("GET", "/getID")
	ts.Exactly(http.StatusNotFound, rr.Code)

	ts.Require().NoError(ts.store.Save("getID", *ts.newAdvancedGame()))

	rr = ts.newRequest("GET", "/getID")
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Dices": [
			{
				"Locked": true,
				"Value": 3
			},
			{
				"Locked": false,
				"Value": 2
			},
			{
				"Locked": true,
				"Value": 3
			},
			{
				"Locked": false,
				"Value": 1
			},
			{
				"Locked": false,
				"Value": 5
			}
		],
		"Players": [
			{
				"User": "Alice",
				"ScoreSheet": {
					"fives": 15,
					"full-house": 25,
					"twos": 6
				}
			},
			{
				"User": "Bob",
				"ScoreSheet": {
					"four-of-a-kind": 16,
					"threes": 6
				}
			},
			{
				"User": "Carol",
				"ScoreSheet": {
					"small-straight": 30,
					"twos": 6
				}
			}
		],
		"Round": 5,
		"CurrentPlayer": 1,
		"RollCount": 1
	}`, rr.Body.String())
}

func (ts *testSuite) newRequest(method string, url string, query ...string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, nil)
	ts.NoError(err)

	if len(query) > 0 {
		q := req.URL.Query()
		for _, kv := range query {
			got := strings.Split(kv, "=")
			ts.Require().Len(got, 2)
			q.Add(got[0], got[1])
		}
		req.URL.RawQuery = q.Encode()
	}

	rr := httptest.NewRecorder()
	ts.handler.ServeHTTP(rr, req)

	return rr
}

func (ts *testSuite) newAdvancedGame() *model.Game {
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
