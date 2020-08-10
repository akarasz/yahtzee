package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akarasz/yahtzee/pkg/models"
	"github.com/akarasz/yahtzee/pkg/store"
)

func TestRootHandler_newGame(t *testing.T) {
	t.Run("should put new game to store", func(t *testing.T) {
		s := &storeStub{}
		h := RootHandler{
			store: s,
			game:  &GameHandler{},
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth("alice", "")

		h.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusCreated; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
		if got, want := s.putCallCount, 1; got != want {
			t.Errorf("wrong number of invocation of store.Put: got %v want %v", got, want)
		}
	})

	t.Run("should fail without basic auth header", func(t *testing.T) {
		h := RootHandler{
			store: &storeStub{},
			game:  &GameHandler{},
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusUnauthorized; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})

	t.Run("should fail when a game already in store with same id", func(t *testing.T) {
		s := &storeStub{
			putError: store.ErrAlreadyExists,
		}
		h := RootHandler{
			store: s,
			game:  &GameHandler{},
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth("alice", "")

		h.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusInternalServerError; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

func TestRootHandler_existingGame(t *testing.T) {
	t.Run("should call game handler with game from store and user from auth", func(t *testing.T) {
		g := &models.Game{}
		s := &storeStub{
			getGame: g,
		}
		gh := &gameHandlerStub{}
		h := RootHandler{
			store: s,
			game:  gh,
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/gameId", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth("alice", "")

		h.ServeHTTP(rr, req)

		if got, want := gh.handleCallCount, 1; got != want {
			t.Fatalf("wrong number of handler invocations. got %v, want %v", got, want)
		}
		if got, want := gh.handleCallParamG[0], g; got != want {
			t.Errorf("wrong game was passed to handler. got %v want %v", got, want)
		}
		if got, want := gh.handleCallParamUser[0], "alice"; got != want {
			t.Errorf("wrong user was passed to handler. got %v want %v", got, want)
		}
	})

	t.Run("should fail without basic auth header", func(t *testing.T) {
		h := RootHandler{
			store: &storeStub{},
			game:  &GameHandler{},
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/gameID", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusUnauthorized; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})

	t.Run("should fail when a game not found in store", func(t *testing.T) {
		s := &storeStub{
			getError: store.ErrNotExists,
		}
		h := RootHandler{
			store: s,
			game:  &GameHandler{},
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/gameId", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth("alice", "")

		h.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusNotFound; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

type storeStub struct {
	putError     error
	putCallCount int

	getGame      *models.Game
	getError     error
	getCallCount int
}

func (s *storeStub) Put(id string, g *models.Game) error {
	s.putCallCount++
	return s.putError
}

func (s *storeStub) Get(id string) (*models.Game, error) {
	s.getCallCount++
	return s.getGame, s.getError
}

type gameHandlerStub struct {
	handleCallCount     int
	handleCallParamG    []*models.Game
	handleCallParamUser []string
}

func (h *gameHandlerStub) handle(user string, g *models.Game) http.Handler {
	h.handleCallCount++
	h.handleCallParamG = append(h.handleCallParamG, g)
	h.handleCallParamUser = append(h.handleCallParamUser, user)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
