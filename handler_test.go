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

type TestHandlerSuite struct {
	suite.Suite

	store *store.InMemory
	event *event.InApp

	handler http.Handler
}

func TestSuite(t *testing.T) {
	s := store.New()
	e := event.New()

	suite.Run(t, &TestHandlerSuite{
		store:   s,
		event:   e,
		handler: yahtzee.NewHandler(s, e, e),
	})
}

func (ts *TestHandlerSuite) TestCreate() {
	rr := ts.newRequest("POST", "/")
	ts.Exactly(http.StatusCreated, rr.Code)
	if ts.Contains(rr.HeaderMap, "Location") && ts.Len(rr.HeaderMap["Location"], 1) {
		got, err := ts.store.Load(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		ts.Require().NoError(err)
		ts.Exactly(*model.NewGame(), got)
	}
}

func (ts *TestHandlerSuite) TestHints() {
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

func (ts *TestHandlerSuite) newRequest(method string, url string, query ...string) *httptest.ResponseRecorder {
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
