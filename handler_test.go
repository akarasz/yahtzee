package yahtzee_test

import (
	"encoding/base64"
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
	ts.store.Save("addPlayer-advancedID", *advanced)

	rr = ts.record(request("POST", "/addPlayer-advancedID/join"), asUser("Alice"))
	ts.Exactly(http.StatusBadRequest, rr.Code)

	// request successful (200)
	game := model.NewGame()
	ts.store.Save("addPlayerID", *game)

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
