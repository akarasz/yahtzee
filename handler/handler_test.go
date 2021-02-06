package handler_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/event"
	event_impl "github.com/akarasz/yahtzee/event/embedded"
	"github.com/akarasz/yahtzee/handler"
	store "github.com/akarasz/yahtzee/store/embedded"
)

type testSuite struct {
	suite.Suite

	store *store.InMemory
	event *event_impl.InApp

	handler http.Handler
}

func TestSuite(t *testing.T) {
	s := store.New()
	e := event_impl.New()

	suite.Run(t, &testSuite{
		store:   s,
		event:   e,
		handler: handler.New(s, e, e),
	})
}

func (ts *testSuite) TestCreate() {
	rr := ts.record(request("POST", "/"))
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		created := ts.fromStore(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		ts.Exactly(yahtzee.NewGame(), created)
	}
}

func (ts *testSuite) TestCreateSixDice() {
	rr := ts.record(request("POST", "/", "[\"six-dice\"]"))
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		created := ts.fromStore(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		ts.Exactly(yahtzee.NewGame(yahtzee.SixDice), created)
	}
}

func (ts *testSuite) TestCreateYahtzeeBonus() {
	rr := ts.record(request("POST", "/", `["yahtzee-bonus"]`))
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		created := ts.fromStore(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		expected := yahtzee.NewGame(yahtzee.YahtzeeBonus)
		ts.Exactly(expected.CurrentPlayer, created.CurrentPlayer)
		ts.Exactly(expected.Players, created.Players)
		ts.Exactly(expected.Dices, created.Dices)
		ts.Exactly(expected.Features, created.Features)
		ts.Exactly(expected.RollCount, created.RollCount)
		ts.Exactly(expected.Round, created.Round)
	}
}

func (ts *testSuite) TestCreateTheChance() {
	rr := ts.record(request("POST", "/", `["the-chance"]`))
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		created := ts.fromStore(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		expected := yahtzee.NewGame(yahtzee.TheChance)
		ts.Exactly(expected.CurrentPlayer, created.CurrentPlayer)
		ts.Exactly(expected.Players, created.Players)
		ts.Exactly(expected.Dices, created.Dices)
		ts.Exactly(expected.Features, created.Features)
		ts.Exactly(expected.RollCount, created.RollCount)
		ts.Exactly(expected.Round, created.Round)
	}
}

func (ts *testSuite) TestHints() {
	badInputs := []struct {
		description string
		key         string
		value       string
	}{
		{"no query", "noop", "true"},
		{"empty dices", "dices", "1,2,3,4"},
		{"too few dices", "dices", "1,2,3,4"},
		{"too many dices", "dices", "1,2,3,4,5,6"},
		{"has low face value", "dices", "1,1,1,0,1"},
		{"has high face value", "dices", "7,6,6,6,6"},
	}
	for _, tc := range badInputs {
		rr := ts.record(request("GET", "/score"), withQuery(tc.key, tc.value))
		ts.Exactly(http.StatusBadRequest, rr.Code, "when %s", tc.description)
	}

	rr := ts.record(request("GET", "/score"), withQuery("dices", "3,2,6,4,5"))
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

func (ts *testSuite) TestHintsYahtzee() {
	rr := ts.record(request("GET", "/score"), withQuery("dices", "5,5,5,5,5"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":0,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":25,
			"sixes":0,
			"three-of-a-kind":15,
			"four-of-a-kind":20,
			"full-house":0,
			"small-straight":0,
			"large-straight":0,
			"yahtzee":50,
			"chance":25
		}`, rr.Body.String())

	rr = ts.record(request("GET", "/score"), withQuery("dices", "6,6,6,6,6"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":0,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":0,
			"sixes":30,
			"three-of-a-kind":18,
			"four-of-a-kind":24,
			"full-house":0,
			"small-straight":0,
			"large-straight":0,
			"yahtzee":50,
			"chance":30
		}`, rr.Body.String())

	rr = ts.record(request("GET", "/score"), withQuery("dices", "1,1,1,1,1"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":5,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":0,
			"sixes":0,
			"three-of-a-kind":3,
			"four-of-a-kind":4,
			"full-house":0,
			"small-straight":0,
			"large-straight":0,
			"yahtzee":50,
			"chance":5
		}`, rr.Body.String())
}

func (ts *testSuite) TestHintsSixDice() {
	badInputs := []struct {
		description string
		key         string
		value       string
	}{
		{"no query", "noop", "true"},
		{"empty dices", "dices", "1,2,3,4"},
		{"too few dices", "dices", "1,2,3,4,5"},
		{"too many dices", "dices", "1,2,3,4,5,6,6"},
		{"has low face value", "dices", "1,1,1,0,1,1"},
		{"has high face value", "dices", "7,6,6,6,6,6"},
	}
	for _, tc := range badInputs {
		rr := ts.record(request("GET", "/score"), withQuery(tc.key, tc.value), withQuery("features", "six-dice"))
		ts.Exactly(http.StatusBadRequest, rr.Code, "when %s", tc.description)
	}

	rr := ts.record(request("GET", "/score"), withQuery("dices", "3,2,6,4,5,1"), withQuery("features", "six-dice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":1,
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

	rr = ts.record(request("GET", "/score"), withQuery("dices", "1,3,2,6,4,5"), withQuery("features", "six-dice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":1,
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

func (ts *testSuite) TestHintsSixDiceYahtzee() {
	rr := ts.record(request("GET", "/score"), withQuery("dices", "5,5,5,5,5,1"), withQuery("features", "six-dice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":1,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":25,
			"sixes":0,
			"three-of-a-kind":15,
			"four-of-a-kind":20,
			"full-house":0,
			"small-straight":0,
			"large-straight":0,
			"yahtzee":50,
			"chance":25
		}`, rr.Body.String())

	rr = ts.record(request("GET", "/score"), withQuery("dices", "5,5,5,5,5,5"), withQuery("features", "six-dice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":0,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":25,
			"sixes":0,
			"three-of-a-kind":15,
			"four-of-a-kind":20,
			"full-house":0,
			"small-straight":0,
			"large-straight":0,
			"yahtzee":50,
			"chance":25
		}`, rr.Body.String())
}

func (ts *testSuite) TestHintsYahtzeeBonus() {
	rr := ts.record(request("GET", "/score"), withQuery("dices", "5,5,5,5,5"), withQuery("features", "yahtzee-bonus"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":0,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":25,
			"sixes":0,
			"three-of-a-kind":15,
			"four-of-a-kind":20,
			"full-house":25,
			"small-straight":30,
			"large-straight":40,
			"yahtzee":50,
			"chance":25
		}`, rr.Body.String())
}

func (ts *testSuite) TestHintsSixDiceYahtzeeBonus() {
	rr := ts.record(request("GET", "/score"), withQuery("dices", "5,5,5,5,5,1"), withQuery("features", "yahtzee-bonus,six-dice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"ones":1,
			"twos":0,
			"threes":0,
			"fours":0,
			"fives":25,
			"sixes":0,
			"three-of-a-kind":15,
			"four-of-a-kind":20,
			"full-house":25,
			"small-straight":30,
			"large-straight":40,
			"yahtzee":50,
			"chance":25
		}`, rr.Body.String())
}

func (ts *testSuite) TestGet() {
	// game not exists
	rr := ts.record(request("GET", "/getID"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// success
	ts.Require().NoError(ts.store.Save("getID", yahtzee.Game{
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
		Features:      []yahtzee.Feature{},
	}))

	rr = ts.record(request("GET", "/getID"))
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
		"RollCount": 1,
		"Features":[]
	}`, rr.Body.String())

	// six-dice
	ts.Require().NoError(ts.store.Save("getID", yahtzee.Game{
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
			{Value: 5, Locked: true},
			{Value: 5, Locked: true},
		},
		Round:         5,
		CurrentPlayer: 1,
		RollCount:     1,
		Features:      []yahtzee.Feature{yahtzee.SixDice},
	}))

	rr = ts.record(request("GET", "/getID"))
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
				"Locked": true,
				"Value": 5
			},
			{
				"Locked": true,
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
		"RollCount": 1,
		"Features":["six-dice"]
	}`, rr.Body.String())
}

func (ts *testSuite) TestAddPlayer() {
	// missing user
	rr := ts.record(request("POST", "/addPlayerID/join"))
	ts.Exactly(http.StatusUnauthorized, rr.Code)

	// game not exists
	rr = ts.record(request("POST", "/addPlayerID/join"), asUser("Alice"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// game already started
	advanced := yahtzee.NewGame()
	advanced.Round = 8
	ts.Require().NoError(ts.store.Save("addPlayer-advancedID", *advanced))

	rr = ts.record(request("POST", "/addPlayer-advancedID/join"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// request successful (200)
	game := yahtzee.NewGame()
	ts.Require().NoError(ts.store.Save("addPlayerID", *game))

	eChan := ts.receiveEvents("addPlayerID")
	rr = ts.record(request("POST", "/addPlayerID/join"), asUser("Alice"))
	ts.Exactly(http.StatusCreated, rr.Code)
	ts.JSONEq(`{
		"Players": [
			{
				"User": "Alice",
				"ScoreSheet": {}
			}
		]
	}`, rr.Body.String())

	// player is saved in store
	saved := ts.fromStore("addPlayerID")
	ts.Exactly(*yahtzee.NewUser("Alice"), saved.Players[0].User)

	// add player event emitted
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.AddPlayer, got.Action)
		ts.Exactly(&handler.AddPlayerResponse{
			Players: []*yahtzee.Player{yahtzee.NewPlayer("Alice")},
		}, got.Data)
	}

	// player already joined
	rr = ts.record(request("POST", "/addPlayerID/join"), asUser("Alice"))
	ts.Exactly(http.StatusConflict, rr.Code)
}

func (ts *testSuite) TestRoll() {
	// missing user
	rr := ts.record(request("POST", "/rollID/roll"))
	ts.Exactly(http.StatusUnauthorized, rr.Code)

	// game not exists
	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// no players yet
	g := yahtzee.NewGame()
	ts.Require().NoError(ts.store.Save("rollID", *g))

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 1
	ts.Require().NoError(ts.store.Save("rollID", *g))

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// game is over
	g.CurrentPlayer = 0
	g.Round = 13
	ts.Require().NoError(ts.store.Save("rollID", *g))

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// out of rolls
	g.Round = 0
	g.RollCount = 3
	ts.Require().NoError(ts.store.Save("rollID", *g))

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// success
	g.Round = 0
	g.RollCount = 0
	ts.Require().NoError(ts.store.Save("rollID", *g))

	eChan := ts.receiveEvents("rollID")

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)

	saved := ts.fromStore("rollID")
	ts.Exactly(g.RollCount+1, saved.RollCount)
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Roll, got.Action)

		ts.Exactly(saved.RollCount, got.Data.(*handler.RollResponse).RollCount)
		ts.Exactly(saved.Dices, got.Data.(*handler.RollResponse).Dices)

		if eventJSON, err := json.Marshal(got.Data.(*handler.RollResponse)); ts.NoError(err) {
			ts.JSONEq(string(eventJSON), rr.Body.String())
		}
	}

	// six-dice
	g.Round = 0
	g.RollCount = 0
	g.Features = []yahtzee.Feature{yahtzee.SixDice}
	ts.Require().NoError(ts.store.Save("rollID", *g))

	eChan = ts.receiveEvents("rollID")

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)

	saved = ts.fromStore("rollID")
	ts.Exactly(g.RollCount+1, saved.RollCount)
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Roll, got.Action)

		ts.Exactly(saved.RollCount, got.Data.(*handler.RollResponse).RollCount)
		ts.Exactly(saved.Dices, got.Data.(*handler.RollResponse).Dices)

		if eventJSON, err := json.Marshal(got.Data.(*handler.RollResponse)); ts.NoError(err) {
			ts.JSONEq(string(eventJSON), rr.Body.String())
		}
	}
}

func (ts *testSuite) TestRollingALot() {
	g := yahtzee.NewGame()
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
	}
	g.Dices[1] = &yahtzee.Dice{
		Value:  3,
		Locked: true,
	}
	g.Dices[4] = &yahtzee.Dice{
		Value:  2,
		Locked: true,
	}
	ts.Require().NoError(ts.store.Save("rollALotID", *g))

	for i := 0; i < 1000; i++ {
		ts.record(request("POST", "/rollALotID/roll"), asUser("Alice"))
		afterRoll := ts.fromStore("rollALotID")
		ts.Require().GreaterOrEqual(afterRoll.Dices[0].Value, 1)
		ts.Require().LessOrEqual(afterRoll.Dices[0].Value, 6)
		ts.Require().Exactly(3, afterRoll.Dices[1].Value)
		ts.Require().GreaterOrEqual(afterRoll.Dices[2].Value, 1)
		ts.Require().LessOrEqual(afterRoll.Dices[2].Value, 6)
		ts.Require().GreaterOrEqual(afterRoll.Dices[3].Value, 1)
		ts.Require().LessOrEqual(afterRoll.Dices[3].Value, 6)
		ts.Require().Exactly(2, afterRoll.Dices[4].Value)
	}
}

func (ts *testSuite) TestLock() {
	// missing user
	rr := ts.record(request("POST", "/lockID/lock/2"))
	ts.Exactly(http.StatusUnauthorized, rr.Code)

	// game not exists
	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// no players yet
	g := yahtzee.NewGame()
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 1
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// game is over
	g.CurrentPlayer = 0
	g.Round = 13
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// no roll happened yet
	g.Round = 0
	g.RollCount = 0
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// out of rolls
	g.RollCount = 3
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// invalid dice index
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/-1"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)
	rr = ts.record(request("POST", "/lockID/lock/5"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// locks an unlocked dice
	ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	saved := ts.fromStore("lockID")
	ts.True(saved.Dices[2].Locked)

	// unlocks a locked dice
	ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	saved = ts.fromStore("lockID")
	ts.False(saved.Dices[2].Locked)

	// successful request
	eChan := ts.receiveEvents("lockID")

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Dices": [
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": true
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			}
		]
	}`, rr.Body.String())

	saved = ts.fromStore("lockID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Lock, got.Action)

		ts.Exactly(saved.Dices, got.Data.(*handler.LockResponse).Dices)
	}

	// six-dice
	g = yahtzee.NewGame(yahtzee.SixDice)
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 0
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("lockID", *g))
	eChan = ts.receiveEvents("lockID")

	rr = ts.record(request("POST", "/lockID/lock/5"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Dices": [
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": true
			}
		]
	}`, rr.Body.String())

	saved = ts.fromStore("lockID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Lock, got.Action)

		ts.Exactly(saved.Dices, got.Data.(*handler.LockResponse).Dices)
	}
}

func (ts *testSuite) TestScore() {
	// missing user
	rr := ts.record(request("POST", "/scoreID/score", "chance"))
	ts.Exactly(http.StatusUnauthorized, rr.Code)

	// game not exists
	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// no players
	g := yahtzee.NewGame()
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// game is over
	g.CurrentPlayer = 0
	g.Round = 13
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// roll first
	g.Round = 0
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// invalid category
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)
	rr = ts.record(request("POST", "/scoreID/score", "wat"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// category is already scored
	g.Players[0].ScoreSheet[yahtzee.FullHouse] = 25
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "full-house"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// successful request
	eChan := ts.receiveEvents("scoreID")

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Players": [
			{
				"User": "Alice",
				"ScoreSheet": {
					"chance": 5,
					"full-house": 25
				}
			},
			{
				"User": "Bob",
				"ScoreSheet": {}
			}
		],
		"Dices": [
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			}
		],
		"Round": 0,
		"CurrentPlayer": 1,
		"RollCount": 0,
		"Features": []
	}`, rr.Body.String())

	saved := ts.fromStore("scoreID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Score, got.Action)
		ts.Exactly(saved, got.Data.(*yahtzee.Game))
	}

	// scoring
	scoringCases := []struct {
		dices    []int
		category yahtzee.Category
		value    int
	}{
		{[]int{1, 2, 3, 1, 1}, yahtzee.Ones, 3},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Twos, 4},
		{[]int{6, 4, 2, 2, 3}, yahtzee.Threes, 3},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Fours, 0},
		{[]int{4, 4, 1, 2, 4}, yahtzee.Fours, 12},
		{[]int{6, 6, 3, 5, 2}, yahtzee.Fives, 5},
		{[]int{5, 3, 6, 6, 6}, yahtzee.Sixes, 18},
		{[]int{2, 4, 3, 6, 4}, yahtzee.ThreeOfAKind, 0},
		{[]int{3, 1, 3, 1, 3}, yahtzee.ThreeOfAKind, 9},
		{[]int{5, 2, 5, 5, 5}, yahtzee.ThreeOfAKind, 15},
		{[]int{2, 6, 3, 2, 2}, yahtzee.FourOfAKind, 0},
		{[]int{1, 6, 6, 6, 6}, yahtzee.FourOfAKind, 24},
		{[]int{4, 4, 4, 4, 4}, yahtzee.FourOfAKind, 16},
		{[]int{5, 5, 2, 5, 5}, yahtzee.FullHouse, 0},
		{[]int{2, 5, 3, 6, 5}, yahtzee.FullHouse, 0},
		{[]int{5, 5, 2, 5, 2}, yahtzee.FullHouse, 25},
		{[]int{3, 1, 3, 1, 3}, yahtzee.FullHouse, 25},
		{[]int{6, 2, 5, 1, 3}, yahtzee.SmallStraight, 0},
		{[]int{6, 2, 4, 1, 3}, yahtzee.SmallStraight, 30},
		{[]int{4, 2, 3, 5, 3}, yahtzee.SmallStraight, 30},
		{[]int{1, 6, 3, 5, 4}, yahtzee.SmallStraight, 30},
		{[]int{3, 5, 2, 3, 4}, yahtzee.LargeStraight, 0},
		{[]int{3, 5, 2, 1, 4}, yahtzee.LargeStraight, 40},
		{[]int{5, 2, 6, 3, 4}, yahtzee.LargeStraight, 40},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Yahtzee, 50},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Yahtzee, 50},
		{[]int{6, 2, 4, 1, 3}, yahtzee.Chance, 16},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Chance, 18},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Chance, 14},
	}

	for _, tc := range scoringCases {
		g := yahtzee.NewGame()
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		ts.Require().NoError(ts.store.Save("score_scoringID", *g))

		ts.record(request("POST", "/score_scoringID/score", string(tc.category)), asUser("Alice"))

		got := ts.fromStore("score_scoringID")
		ts.Exactly(tc.value, got.Players[0].ScoreSheet[tc.category],
			"should return %d for %q on %v", tc.value, tc.category, tc.dices)
	}

	// bonus
	bonusCases := []struct {
		dices         []int
		upperSection  []int
		scoring       yahtzee.Category
		givesBonus    bool
		mustHaveValue bool
	}{
		{[]int{1, 3, 6, 2, 4}, []int{3, 6, -1, 16, 25, -1}, yahtzee.Sixes, false, false},
		{[]int{1, 3, 6, 2, 4}, []int{-1, -1, 12, -1, 20, 36}, yahtzee.Fours, true, false},
		{[]int{1, 3, 6, 2, 4}, []int{3, 6, 9, 16, 25, -1}, yahtzee.Sixes, true, true},
		{[]int{1, 1, 3, 3, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, false, true},
		{[]int{1, 1, 1, 3, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
		{[]int{1, 1, 1, 1, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
	}

	for _, tc := range bonusCases {
		g := yahtzee.NewGame()
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		if tc.upperSection[0] > 0 {
			g.Players[0].ScoreSheet["ones"] = tc.upperSection[0]
		}
		if tc.upperSection[1] > 0 {
			g.Players[0].ScoreSheet["twos"] = tc.upperSection[1]
		}
		if tc.upperSection[2] > 0 {
			g.Players[0].ScoreSheet["threes"] = tc.upperSection[2]
		}
		if tc.upperSection[3] > 0 {
			g.Players[0].ScoreSheet["fours"] = tc.upperSection[3]
		}
		if tc.upperSection[4] > 0 {
			g.Players[0].ScoreSheet["fives"] = tc.upperSection[4]
		}
		if tc.upperSection[5] > 0 {
			g.Players[0].ScoreSheet["sixes"] = tc.upperSection[5]
		}
		ts.Require().NoError(ts.store.Save("score_bonusID", *g))

		rr := ts.record(request("POST", "/score_bonusID/score", string(tc.scoring)), asUser("Alice"))

		got := ts.fromStore("score_bonusID")
		bonus, hasBonus := got.Players[0].ScoreSheet["bonus"]
		if tc.mustHaveValue {
			ts.True(hasBonus)
		}

		if tc.givesBonus {
			ts.Exactly(35, bonus, "should have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		} else {
			ts.Exactly(0, bonus, "should not have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		}
	}

	// counters
	counterCases := []struct {
		description string

		round         int
		currentPlayer int
		rollCount     int

		nextRound         int
		nextCurrentPlayer int
		nextRollCount     int
	}{
		{"mid-round after one roll", 0, 0, 1, 0, 1, 0},
		{"mid-round after two rolls", 0, 0, 2, 0, 1, 0},
		{"mid-round for last player", 0, 1, 2, 1, 0, 0},
	}

	for _, tc := range counterCases {
		g := yahtzee.NewGame()
		g.Players = []*yahtzee.Player{
			yahtzee.NewPlayer("Alice"),
			yahtzee.NewPlayer("Bob"),
		}
		g.Round = tc.round
		g.CurrentPlayer = tc.currentPlayer
		g.RollCount = tc.rollCount

		ts.Require().NoError(ts.store.Save("score_counterID", *g))

		ts.record(
			request("POST", "/score_counterID/score", "chance"),
			asUser(string(g.Players[tc.currentPlayer].User)))

		got := ts.fromStore("score_counterID")
		ts.Exactly(tc.nextRound, got.Round, "for %s", tc.description)
		ts.Exactly(tc.nextCurrentPlayer, got.CurrentPlayer, "for %s", tc.description)
		ts.Exactly(tc.nextRollCount, got.RollCount, "for %s", tc.description)
	}
}

func (ts *testSuite) TestScoreTheChance() {
	// Last round, TheChance enabled
	g := yahtzee.NewGame(yahtzee.TheChance)
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.Players[1].ScoreSheet[yahtzee.Ones] = 1
	g.CurrentPlayer = 0
	g.Round = 12
	g.RollCount = 1

	ts.Require().NoError(ts.store.Save("scoreID", *g))
	eChan := ts.receiveEvents("scoreID")

	rr := ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Players": [
			{
				"User": "Alice",
				"ScoreSheet": {
					"chance": 5
				}
			},
			{
				"User": "Bob",
				"ScoreSheet": {
					"ones": 1
				}
			}
		],
		"Dices": [
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			}
		],
		"Round": 12,
		"CurrentPlayer": 1,
		"RollCount": 0,
		"Features": ["the-chance"]
	}`, rr.Body.String())

	saved := ts.fromStore("scoreID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Score, got.Action)
		ts.Exactly(saved, got.Data.(*yahtzee.Game))
	}

	saved.RollCount = 1
	ts.Require().NoError(ts.store.Save("scoreID", *saved))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Bob"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
		"Players": [
			{
				"User": "Alice",
				"ScoreSheet": {
					"chance": 5,
					"chance-bonus": 495
				}
			},
			{
				"User": "Bob",
				"ScoreSheet": {
					"ones": 1,
					"chance": 5
				}
			}
		],
		"Dices": [
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			},
			{
				"Value": 1,
				"Locked": false
			}
		],
		"Round": 13,
		"CurrentPlayer": 0,
		"RollCount": 0,
		"Features": ["the-chance"]
	}`, rr.Body.String())

	saved = ts.fromStore("scoreID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Score, got.Action)
		ts.Exactly(saved, got.Data.(*yahtzee.Game))
	}

}

func (ts *testSuite) TestScoreSixDice() {
	// no players
	g := yahtzee.NewGame(yahtzee.SixDice)
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr := ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// game is over
	g.CurrentPlayer = 0
	g.Round = 13
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// roll first
	g.Round = 0
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// invalid category
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)
	rr = ts.record(request("POST", "/scoreID/score", "wat"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// category is already scored
	g.Players[0].ScoreSheet[yahtzee.FullHouse] = 25
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "full-house"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// successful request
	eChan := ts.receiveEvents("scoreID")

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"Players": [
				{
					"User": "Alice",
					"ScoreSheet": {
						"chance": 5,
						"full-house": 25
					}
				},
				{
					"User": "Bob",
					"ScoreSheet": {}
				}
			],
			"Dices": [
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				}
			],
			"Round": 0,
			"CurrentPlayer": 1,
			"RollCount": 0,
			"Features": ["six-dice"]
		}`, rr.Body.String())

	saved := ts.fromStore("scoreID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Score, got.Action)
		ts.Exactly(saved, got.Data.(*yahtzee.Game))
	}

	// scoring
	scoringCases := []struct {
		dices    []int
		category yahtzee.Category
		value    int
	}{
		{[]int{1, 2, 3, 1, 1, 4}, yahtzee.Ones, 3},
		{[]int{2, 3, 4, 2, 3, 5}, yahtzee.Twos, 4},
		{[]int{6, 4, 2, 2, 3, 5}, yahtzee.Threes, 3},
		{[]int{1, 6, 3, 3, 5, 2}, yahtzee.Fours, 0},
		{[]int{4, 4, 1, 2, 4, 5}, yahtzee.Fours, 12},
		{[]int{6, 6, 3, 5, 2, 1}, yahtzee.Fives, 5},
		{[]int{5, 3, 6, 6, 6, 3}, yahtzee.Sixes, 18},
		{[]int{2, 4, 3, 6, 4, 1}, yahtzee.ThreeOfAKind, 0},
		{[]int{3, 1, 3, 1, 3, 1}, yahtzee.ThreeOfAKind, 9},
		{[]int{5, 2, 5, 5, 5, 2}, yahtzee.ThreeOfAKind, 15},
		{[]int{2, 6, 3, 2, 2, 1}, yahtzee.FourOfAKind, 0},
		{[]int{1, 6, 6, 6, 6, 2}, yahtzee.FourOfAKind, 24},
		{[]int{4, 4, 4, 4, 4, 3}, yahtzee.FourOfAKind, 16},
		{[]int{5, 5, 2, 5, 5, 3}, yahtzee.FullHouse, 0},
		{[]int{2, 5, 3, 6, 5, 3}, yahtzee.FullHouse, 0},
		{[]int{5, 5, 2, 5, 2, 2}, yahtzee.FullHouse, 25},
		{[]int{3, 1, 3, 1, 3, 3}, yahtzee.FullHouse, 25},
		{[]int{6, 2, 5, 1, 3, 2}, yahtzee.SmallStraight, 0},
		{[]int{6, 2, 4, 1, 3, 4}, yahtzee.SmallStraight, 30},
		{[]int{4, 2, 3, 5, 3, 3}, yahtzee.SmallStraight, 30},
		{[]int{1, 6, 3, 5, 4, 2}, yahtzee.SmallStraight, 30},
		{[]int{3, 5, 2, 3, 4, 4}, yahtzee.LargeStraight, 0},
		{[]int{3, 5, 2, 1, 4, 6}, yahtzee.LargeStraight, 40},
		{[]int{5, 2, 6, 3, 4, 3}, yahtzee.LargeStraight, 40},
		{[]int{3, 3, 3, 3, 3, 1}, yahtzee.Yahtzee, 50},
		{[]int{1, 1, 1, 1, 1, 1}, yahtzee.Yahtzee, 50},
		{[]int{6, 2, 4, 1, 3, 1}, yahtzee.Chance, 16},
		{[]int{1, 6, 3, 3, 5, 1}, yahtzee.Chance, 18},
		{[]int{2, 3, 4, 2, 3, 2}, yahtzee.Chance, 14},
	}

	for _, tc := range scoringCases {
		g := yahtzee.NewGame(yahtzee.SixDice)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 6; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		ts.Require().NoError(ts.store.Save("score_scoringID", *g))

		ts.record(request("POST", "/score_scoringID/score", string(tc.category)), asUser("Alice"))

		got := ts.fromStore("score_scoringID")
		ts.Exactly(tc.value, got.Players[0].ScoreSheet[tc.category],
			"should return %d for %q on %v", tc.value, tc.category, tc.dices)
	}

	// bonus
	bonusCases := []struct {
		dices         []int
		upperSection  []int
		scoring       yahtzee.Category
		givesBonus    bool
		mustHaveValue bool
	}{
		{[]int{1, 3, 6, 2, 4, 1}, []int{3, 6, -1, 16, 25, -1}, yahtzee.Sixes, false, false},
		{[]int{1, 3, 6, 2, 4, 3}, []int{-1, -1, 12, -1, 20, 36}, yahtzee.Fours, true, false},
		{[]int{1, 3, 6, 2, 4, 1}, []int{3, 6, 9, 16, 25, -1}, yahtzee.Sixes, true, true},
		{[]int{1, 1, 3, 3, 3, 2}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, false, true},
		{[]int{1, 1, 1, 3, 3, 2}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
		{[]int{1, 1, 1, 1, 3, 2}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
	}

	for _, tc := range bonusCases {
		g := yahtzee.NewGame(yahtzee.SixDice)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 6; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		if tc.upperSection[0] > 0 {
			g.Players[0].ScoreSheet["ones"] = tc.upperSection[0]
		}
		if tc.upperSection[1] > 0 {
			g.Players[0].ScoreSheet["twos"] = tc.upperSection[1]
		}
		if tc.upperSection[2] > 0 {
			g.Players[0].ScoreSheet["threes"] = tc.upperSection[2]
		}
		if tc.upperSection[3] > 0 {
			g.Players[0].ScoreSheet["fours"] = tc.upperSection[3]
		}
		if tc.upperSection[4] > 0 {
			g.Players[0].ScoreSheet["fives"] = tc.upperSection[4]
		}
		if tc.upperSection[5] > 0 {
			g.Players[0].ScoreSheet["sixes"] = tc.upperSection[5]
		}
		ts.Require().NoError(ts.store.Save("score_bonusID", *g))

		rr := ts.record(request("POST", "/score_bonusID/score", string(tc.scoring)), asUser("Alice"))

		got := ts.fromStore("score_bonusID")
		bonus, hasBonus := got.Players[0].ScoreSheet["bonus"]
		if tc.mustHaveValue {
			ts.True(hasBonus)
		}

		if tc.givesBonus {
			ts.Exactly(35, bonus, "should have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		} else {
			ts.Exactly(0, bonus, "should not have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		}
	}
}

func (ts *testSuite) TestScoreYahtzeeBonus() {
	// no players
	g := yahtzee.NewGame(yahtzee.YahtzeeBonus)
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr := ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*yahtzee.Player{
		yahtzee.NewPlayer("Alice"),
		yahtzee.NewPlayer("Bob"),
	}
	g.CurrentPlayer = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// game is over
	g.CurrentPlayer = 0
	g.Round = 13
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// roll first
	g.Round = 0
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// invalid category
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)
	rr = ts.record(request("POST", "/scoreID/score", "wat"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// category is already scored
	g.Players[0].ScoreSheet[yahtzee.FullHouse] = 25
	ts.Require().NoError(ts.store.Save("scoreID", *g))

	rr = ts.record(request("POST", "/scoreID/score", "full-house"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// successful request
	eChan := ts.receiveEvents("scoreID")

	rr = ts.record(request("POST", "/scoreID/score", "chance"), asUser("Alice"))
	ts.Exactly(http.StatusOK, rr.Code)
	ts.JSONEq(`{
			"Players": [
				{
					"User": "Alice",
					"ScoreSheet": {
						"chance": 5,
						"full-house": 25
					}
				},
				{
					"User": "Bob",
					"ScoreSheet": {}
				}
			],
			"Dices": [
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				},
				{
					"Value": 1,
					"Locked": false
				}
			],
			"Round": 0,
			"CurrentPlayer": 1,
			"RollCount": 0,
			"Features": ["yahtzee-bonus"]
		}`, rr.Body.String())

	saved := ts.fromStore("scoreID")
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.Score, got.Action)
		ts.Exactly(saved, got.Data.(*yahtzee.Game))
	}

	// scoring
	scoringCases := []struct {
		dices    []int
		category yahtzee.Category
		value    int
	}{
		{[]int{1, 2, 3, 1, 1}, yahtzee.Ones, 3},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Twos, 4},
		{[]int{6, 4, 2, 2, 3}, yahtzee.Threes, 3},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Fours, 0},
		{[]int{4, 4, 1, 2, 4}, yahtzee.Fours, 12},
		{[]int{6, 6, 3, 5, 2}, yahtzee.Fives, 5},
		{[]int{5, 3, 6, 6, 6}, yahtzee.Sixes, 18},
		{[]int{2, 4, 3, 6, 4}, yahtzee.ThreeOfAKind, 0},
		{[]int{3, 1, 3, 1, 3}, yahtzee.ThreeOfAKind, 9},
		{[]int{5, 2, 5, 5, 5}, yahtzee.ThreeOfAKind, 15},
		{[]int{2, 6, 3, 2, 2}, yahtzee.FourOfAKind, 0},
		{[]int{1, 6, 6, 6, 6}, yahtzee.FourOfAKind, 24},
		{[]int{4, 4, 4, 4, 4}, yahtzee.FourOfAKind, 16},
		{[]int{5, 5, 2, 5, 5}, yahtzee.FullHouse, 0},
		{[]int{2, 5, 3, 6, 5}, yahtzee.FullHouse, 0},
		{[]int{5, 5, 2, 5, 2}, yahtzee.FullHouse, 25},
		{[]int{3, 1, 3, 1, 3}, yahtzee.FullHouse, 25},
		{[]int{6, 2, 5, 1, 3}, yahtzee.SmallStraight, 0},
		{[]int{6, 2, 4, 1, 3}, yahtzee.SmallStraight, 30},
		{[]int{4, 2, 3, 5, 3}, yahtzee.SmallStraight, 30},
		{[]int{1, 6, 3, 5, 4}, yahtzee.SmallStraight, 30},
		{[]int{3, 5, 2, 3, 4}, yahtzee.LargeStraight, 0},
		{[]int{3, 5, 2, 1, 4}, yahtzee.LargeStraight, 40},
		{[]int{5, 2, 6, 3, 4}, yahtzee.LargeStraight, 40},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Yahtzee, 50},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Yahtzee, 50},
		{[]int{6, 2, 4, 1, 3}, yahtzee.Chance, 16},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Chance, 18},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Chance, 14},
	}

	for _, tc := range scoringCases {
		g := yahtzee.NewGame(yahtzee.YahtzeeBonus)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		ts.Require().NoError(ts.store.Save("score_scoringID", *g))

		ts.record(request("POST", "/score_scoringID/score", string(tc.category)), asUser("Alice"))

		got := ts.fromStore("score_scoringID")
		ts.Exactly(tc.value, got.Players[0].ScoreSheet[tc.category],
			"should return %d for %q on %v", tc.value, tc.category, tc.dices)
	}

	// bonus
	bonusCases := []struct {
		dices         []int
		upperSection  []int
		scoring       yahtzee.Category
		givesBonus    bool
		mustHaveValue bool
	}{
		{[]int{1, 3, 6, 2, 4}, []int{3, 6, -1, 16, 25, -1}, yahtzee.Sixes, false, false},
		{[]int{1, 3, 6, 2, 4}, []int{-1, -1, 12, -1, 20, 36}, yahtzee.Fours, true, false},
		{[]int{1, 3, 6, 2, 4}, []int{3, 6, 9, 16, 25, -1}, yahtzee.Sixes, true, true},
		{[]int{1, 1, 3, 3, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, false, true},
		{[]int{1, 1, 1, 3, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
		{[]int{1, 1, 1, 1, 3}, []int{-1, 2, 3, 4, 15, 36}, yahtzee.Ones, true, true},
	}

	for _, tc := range bonusCases {
		g := yahtzee.NewGame(yahtzee.YahtzeeBonus)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		if tc.upperSection[0] > 0 {
			g.Players[0].ScoreSheet["ones"] = tc.upperSection[0]
		}
		if tc.upperSection[1] > 0 {
			g.Players[0].ScoreSheet["twos"] = tc.upperSection[1]
		}
		if tc.upperSection[2] > 0 {
			g.Players[0].ScoreSheet["threes"] = tc.upperSection[2]
		}
		if tc.upperSection[3] > 0 {
			g.Players[0].ScoreSheet["fours"] = tc.upperSection[3]
		}
		if tc.upperSection[4] > 0 {
			g.Players[0].ScoreSheet["fives"] = tc.upperSection[4]
		}
		if tc.upperSection[5] > 0 {
			g.Players[0].ScoreSheet["sixes"] = tc.upperSection[5]
		}
		ts.Require().NoError(ts.store.Save("score_bonusID", *g))

		rr := ts.record(request("POST", "/score_bonusID/score", string(tc.scoring)), asUser("Alice"))

		got := ts.fromStore("score_bonusID")
		bonus, hasBonus := got.Players[0].ScoreSheet["bonus"]
		if tc.mustHaveValue {
			ts.True(hasBonus)
		}

		if tc.givesBonus {
			ts.Exactly(35, bonus, "should have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		} else {
			ts.Exactly(0, bonus, "should not have bonus for %v when scoring %q", rr.Body.String(), tc.scoring)
		}
	}
}

func (ts *testSuite) TestScoreYahtzeeBonusWhenYahtzeeAlreadyScored() {
	// scoring
	scoringCases := []struct {
		dices    []int
		category yahtzee.Category
		value    int
		yahtzee  int
	}{
		{[]int{1, 2, 3, 1, 1}, yahtzee.Ones, 3, 50},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Ones, 5, 150},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Twos, 4, 50},
		{[]int{2, 2, 2, 2, 2}, yahtzee.Twos, 10, 150},
		{[]int{6, 4, 2, 2, 3}, yahtzee.Threes, 3, 50},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Threes, 15, 150},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Fours, 0, 50},
		{[]int{4, 4, 1, 2, 4}, yahtzee.Fours, 12, 50},
		{[]int{4, 4, 4, 4, 4}, yahtzee.Fours, 20, 150},
		{[]int{6, 6, 3, 5, 2}, yahtzee.Fives, 5, 50},
		{[]int{5, 5, 5, 5, 5}, yahtzee.Fives, 25, 150},
		{[]int{5, 3, 6, 6, 6}, yahtzee.Sixes, 18, 50},
		{[]int{6, 6, 6, 6, 6}, yahtzee.Sixes, 30, 150},
		{[]int{2, 4, 3, 6, 4}, yahtzee.ThreeOfAKind, 0, 50},
		{[]int{3, 1, 3, 1, 3}, yahtzee.ThreeOfAKind, 9, 50},
		{[]int{5, 2, 5, 5, 5}, yahtzee.ThreeOfAKind, 15, 50},
		{[]int{5, 5, 5, 5, 5}, yahtzee.ThreeOfAKind, 15, 150},
		{[]int{2, 6, 3, 2, 2}, yahtzee.FourOfAKind, 0, 50},
		{[]int{1, 6, 6, 6, 6}, yahtzee.FourOfAKind, 24, 50},
		{[]int{4, 4, 4, 4, 4}, yahtzee.FourOfAKind, 16, 150},
		{[]int{5, 5, 2, 5, 5}, yahtzee.FullHouse, 0, 50},
		{[]int{2, 5, 3, 6, 5}, yahtzee.FullHouse, 0, 50},
		{[]int{5, 5, 2, 5, 2}, yahtzee.FullHouse, 25, 50},
		{[]int{3, 1, 3, 1, 3}, yahtzee.FullHouse, 25, 50},
		{[]int{3, 3, 3, 3, 3}, yahtzee.FullHouse, 25, 150},
		{[]int{6, 2, 5, 1, 3}, yahtzee.SmallStraight, 0, 50},
		{[]int{6, 2, 4, 1, 3}, yahtzee.SmallStraight, 30, 50},
		{[]int{4, 2, 3, 5, 3}, yahtzee.SmallStraight, 30, 50},
		{[]int{1, 6, 3, 5, 4}, yahtzee.SmallStraight, 30, 50},
		{[]int{5, 5, 5, 5, 5}, yahtzee.SmallStraight, 30, 150},
		{[]int{3, 5, 2, 3, 4}, yahtzee.LargeStraight, 0, 50},
		{[]int{3, 5, 2, 1, 4}, yahtzee.LargeStraight, 40, 50},
		{[]int{5, 2, 6, 3, 4}, yahtzee.LargeStraight, 40, 50},
		{[]int{5, 5, 5, 5, 5}, yahtzee.LargeStraight, 40, 150},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Yahtzee, 50, 50},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Yahtzee, 50, 50},
		{[]int{6, 2, 4, 1, 3}, yahtzee.Chance, 16, 50},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Chance, 18, 50},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Chance, 14, 50},
		{[]int{2, 2, 2, 2, 2}, yahtzee.Chance, 10, 150},
	}

	for _, tc := range scoringCases {
		g := yahtzee.NewGame(yahtzee.YahtzeeBonus)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		g.Players[0].ScoreSheet[yahtzee.Yahtzee] = 50
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		ts.Require().NoError(ts.store.Save("score_scoringID", *g))

		ts.record(request("POST", "/score_scoringID/score", string(tc.category)), asUser("Alice"))

		got := ts.fromStore("score_scoringID")
		ts.Exactly(tc.value, got.Players[0].ScoreSheet[tc.category],
			"should return %d for %q on %v", tc.value, tc.category, tc.dices)

		ts.Exactly(tc.yahtzee, got.Players[0].ScoreSheet[yahtzee.Yahtzee],
			"should return %d for yahtzee for %q on %v", tc.yahtzee, tc.category, tc.dices)
	}
}

func (ts *testSuite) TestScoreYahtzeeBonusWhenYahtzeeAlreadyZeroed() {
	// scoring
	scoringCases := []struct {
		dices    []int
		category yahtzee.Category
		value    int
		yahtzee  int
	}{
		{[]int{1, 2, 3, 1, 1}, yahtzee.Ones, 3, 0},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Ones, 5, 0},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Twos, 4, 0},
		{[]int{2, 2, 2, 2, 2}, yahtzee.Twos, 10, 0},
		{[]int{6, 4, 2, 2, 3}, yahtzee.Threes, 3, 0},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Threes, 15, 0},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Fours, 0, 0},
		{[]int{4, 4, 1, 2, 4}, yahtzee.Fours, 12, 0},
		{[]int{4, 4, 4, 4, 4}, yahtzee.Fours, 20, 0},
		{[]int{6, 6, 3, 5, 2}, yahtzee.Fives, 5, 0},
		{[]int{5, 5, 5, 5, 5}, yahtzee.Fives, 25, 0},
		{[]int{5, 3, 6, 6, 6}, yahtzee.Sixes, 18, 0},
		{[]int{6, 6, 6, 6, 6}, yahtzee.Sixes, 30, 0},
		{[]int{2, 4, 3, 6, 4}, yahtzee.ThreeOfAKind, 0, 0},
		{[]int{3, 1, 3, 1, 3}, yahtzee.ThreeOfAKind, 9, 0},
		{[]int{5, 2, 5, 5, 5}, yahtzee.ThreeOfAKind, 15, 0},
		{[]int{5, 5, 5, 5, 5}, yahtzee.ThreeOfAKind, 15, 0},
		{[]int{2, 6, 3, 2, 2}, yahtzee.FourOfAKind, 0, 0},
		{[]int{1, 6, 6, 6, 6}, yahtzee.FourOfAKind, 24, 0},
		{[]int{4, 4, 4, 4, 4}, yahtzee.FourOfAKind, 16, 0},
		{[]int{5, 5, 2, 5, 5}, yahtzee.FullHouse, 0, 0},
		{[]int{2, 5, 3, 6, 5}, yahtzee.FullHouse, 0, 0},
		{[]int{5, 5, 2, 5, 2}, yahtzee.FullHouse, 25, 0},
		{[]int{3, 1, 3, 1, 3}, yahtzee.FullHouse, 25, 0},
		{[]int{3, 3, 3, 3, 3}, yahtzee.FullHouse, 25, 0},
		{[]int{6, 2, 5, 1, 3}, yahtzee.SmallStraight, 0, 0},
		{[]int{6, 2, 4, 1, 3}, yahtzee.SmallStraight, 30, 0},
		{[]int{4, 2, 3, 5, 3}, yahtzee.SmallStraight, 30, 0},
		{[]int{1, 6, 3, 5, 4}, yahtzee.SmallStraight, 30, 0},
		{[]int{5, 5, 5, 5, 5}, yahtzee.SmallStraight, 30, 0},
		{[]int{3, 5, 2, 3, 4}, yahtzee.LargeStraight, 0, 0},
		{[]int{3, 5, 2, 1, 4}, yahtzee.LargeStraight, 40, 0},
		{[]int{5, 2, 6, 3, 4}, yahtzee.LargeStraight, 40, 0},
		{[]int{5, 5, 5, 5, 5}, yahtzee.LargeStraight, 40, 0},
		{[]int{3, 3, 3, 3, 3}, yahtzee.Yahtzee, 0, 0},
		{[]int{1, 1, 1, 1, 1}, yahtzee.Yahtzee, 0, 0},
		{[]int{6, 2, 4, 1, 3}, yahtzee.Chance, 16, 0},
		{[]int{1, 6, 3, 3, 5}, yahtzee.Chance, 18, 0},
		{[]int{2, 3, 4, 2, 3}, yahtzee.Chance, 14, 0},
		{[]int{2, 2, 2, 2, 2}, yahtzee.Chance, 10, 0},
	}

	for _, tc := range scoringCases {
		g := yahtzee.NewGame(yahtzee.YahtzeeBonus)
		g.Players = append(g.Players, yahtzee.NewPlayer("Alice"))
		g.RollCount = 1
		g.Players[0].ScoreSheet[yahtzee.Yahtzee] = 0
		for d := 0; d < 5; d++ {
			g.Dices[d].Value = tc.dices[d]
		}
		ts.Require().NoError(ts.store.Save("score_scoringID", *g))

		ts.record(request("POST", "/score_scoringID/score", string(tc.category)), asUser("Alice"))

		got := ts.fromStore("score_scoringID")
		ts.Exactly(tc.value, got.Players[0].ScoreSheet[tc.category],
			"should return %d for %q on %v", tc.value, tc.category, tc.dices)

		ts.Exactly(tc.yahtzee, got.Players[0].ScoreSheet[yahtzee.Yahtzee],
			"should return %d for yahtzee for %q on %v", tc.yahtzee, tc.category, tc.dices)
	}
}

func (ts *testSuite) TestWS() {
	server := httptest.NewServer(ts.handler)
	defer server.Close()
	baseUrl := "ws" + strings.TrimPrefix(server.URL, "http")

	ts.Require().NoError(ts.store.Save("wsID", *yahtzee.NewGame()))

	ws, _, err := websocket.DefaultDialer.Dial(baseUrl+"/wsID/ws", nil)
	if !ts.NoError(err) {
		return
	}
	defer ws.Close()

	ts.event.Emit("wsID", yahtzee.NewUser("Alice"), event.AddPlayer, nil)

	_, p, err := ws.ReadMessage()
	if ts.NoError(err) {
		ts.JSONEq(`{
				"User": "Alice",
				"Action": "add-player",
				"Data": null
			}`, string(p))
	}
}

func (ts *testSuite) TestFeatures() {
	rr := ts.record(request("GET", "/features"))
	ts.Exactly(http.StatusOK, rr.Code)
	features, _ := json.Marshal(yahtzee.Features())
	ts.JSONEq(string(features), rr.Body.String())
}

func (ts *testSuite) record(
	req *http.Request,
	modifiers ...func(*http.Request) *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	for _, modifier := range modifiers {
		req = modifier(req)
	}

	ts.handler.ServeHTTP(rr, req)

	return rr
}

func (ts *testSuite) fromStore(id string) *yahtzee.Game {
	res, err := ts.store.Load(id)
	ts.Require().NoError(err)
	return &res
}

func (ts *testSuite) receiveEvents(id string) chan *event.Event {
	c, err := ts.event.Subscribe(id, id)
	ts.Require().NoError(err)

	res := make(chan *event.Event, 1)

	go func() {
		for {
			select {
			case got := <-c:
				res <- got
			case <-time.After(500 * time.Millisecond):
				res <- nil
				return
			}
		}
	}()

	return res
}

func request(method string, url string, body ...interface{}) *http.Request {
	var bodyReader io.Reader

	if len(body) == 1 {
		switch body[0].(type) {
		case string:
			bodyReader = strings.NewReader(body[0].(string))
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		panic(err)
	}
	return req
}

func withQuery(key, value string) func(*http.Request) *http.Request {
	return func(req *http.Request) *http.Request {
		q := req.URL.Query()
		q.Add(key, value)
		req.URL.RawQuery = q.Encode()
		return req
	}
}

func asUser(name string) func(*http.Request) *http.Request {
	return func(req *http.Request) *http.Request {
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(name+":")))
		return req
	}
}
