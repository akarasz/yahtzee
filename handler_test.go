package yahtzee_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/event"
	event_impl "github.com/akarasz/yahtzee/event/embedded"
	"github.com/akarasz/yahtzee/model"
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
		handler: yahtzee.NewHandler(s, e, e),
	})
}

func (ts *testSuite) TestCreate() {
	rr := ts.record(request("POST", "/"))
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		created := ts.fromStore(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		ts.Exactly(model.NewGame(), created)
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

func (ts *testSuite) TestGet() {
	// game not exists
	rr := ts.record(request("GET", "/getID"))
	ts.Exactly(http.StatusNotFound, rr.Code)

	// success
	ts.Require().NoError(ts.store.Save("getID", model.Game{
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
		"RollCount": 1
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
	advanced := model.NewGame()
	advanced.Round = 8
	ts.Require().NoError(ts.store.Save("addPlayer-advancedID", *advanced))

	rr = ts.record(request("POST", "/addPlayer-advancedID/join"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// request successful (200)
	game := model.NewGame()
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
	ts.Exactly(*model.NewUser("Alice"), saved.Players[0].User)

	// add player event emitted
	if got := <-eChan; ts.NotNil(got) {
		ts.Exactly(event.AddPlayer, got.Action)
		ts.Exactly(&yahtzee.AddPlayerResponse{
			Players: []*model.Player{model.NewPlayer("Alice")},
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
	g := model.NewGame()
	ts.Require().NoError(ts.store.Save("rollID", *g))

	rr = ts.record(request("POST", "/rollID/roll"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*model.Player{
		model.NewPlayer("Alice"),
		model.NewPlayer("Bob"),
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

		ts.Exactly(saved.RollCount, got.Data.(*yahtzee.RollResponse).RollCount)
		ts.Exactly(saved.Dices, got.Data.(*yahtzee.RollResponse).Dices)

		if eventJSON, err := json.Marshal(got.Data.(*yahtzee.RollResponse)); ts.NoError(err) {
			ts.JSONEq(string(eventJSON), rr.Body.String())
		}
	}
}

func (ts *testSuite) TestRollingALot() {
	g := model.NewGame()
	g.Players = []*model.Player{
		model.NewPlayer("Alice"),
	}
	g.Dices[1] = &model.Dice{
		Value:  3,
		Locked: true,
	}
	g.Dices[4] = &model.Dice{
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
	g := model.NewGame()
	g.RollCount = 1
	ts.Require().NoError(ts.store.Save("lockID", *g))

	rr = ts.record(request("POST", "/lockID/lock/2"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// another player's turn
	g.Players = []*model.Player{
		model.NewPlayer("Alice"),
		model.NewPlayer("Bob"),
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

		ts.Exactly(saved.Dices, got.Data.(*yahtzee.LockResponse).Dices)
	}
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

func (ts *testSuite) fromStore(id string) *model.Game {
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

func request(method string, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
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
