package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akarasz/yahtzee/pkg/models"
)

func TestGameHandler_root(t *testing.T) {
	t.Run("should return the game in json", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
		gJson, _ := json.Marshal(g)
		if got, want := strings.TrimSpace(rr.Body.String()), string(gJson); got != want {
			t.Errorf("wrong body: got %q want %q", got, want)
		}
	})
}

func TestGameHandler_join(t *testing.T) {
	t.Run("should return 201", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/join", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusCreated; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})

	t.Run("should add player", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/join", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := c.AddPlayerCallCount, 1; got != want {
			t.Fatalf("wrong number of AddPlayer calls: got %v want %v", got, want)
		}
		if got, want := c.AddPlayerCallParamG[0], g; got != want {
			t.Fatalf("wrong g param for AddPlayer call: got %v want %v", got, want)
		}
		if got, want := c.AddPlayerCallParamName[0], "alice"; got != want {
			t.Fatalf("wrong name param for AddPlayer call: got %v want %v", got, want)
		}
	})

	t.Run("should return 400 when controller returns error", func(t *testing.T) {
		c := &controllerStub{
			AddPlayerReturns: errors.New("random error"),
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/join", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusBadRequest; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

func TestGameHandler_lock(t *testing.T) {
	t.Run("should return the dices", func(t *testing.T) {
		want := []*models.Dice{
			&models.Dice{2, false},
		}
		c := &controllerStub{
			ToggleReturnsDice: want,
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		g.RollCount = 2
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/lock/2", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
		wantJSON, _ := json.Marshal(want)
		if got, want := strings.TrimSpace(rr.Body.String()), string(wantJSON); got != want {
			t.Errorf("wrong body: got %q want %q", got, want)
		}
	})

	t.Run("should call toggle on controller", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/lock/2", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := c.ToggleCallCount, 1; got != want {
			t.Fatalf("wrong number of calls: got %v want %v", got, want)
		}
		if got, want := c.ToggleCallParamG[0], g; got != want {
			t.Fatalf("wrong g param for call: got %v want %v", got, want)
		}
		if got, want := c.ToggleCallParamPlayer[0], "alice"; got != want {
			t.Fatalf("wrong player param for call: got %v want %v", got, want)
		}
		if got, want := c.ToggleCallParamDiceIndex[0], 2; got != want {
			t.Fatalf("wrong diceIndex param for call: got %v want %v", got, want)
		}
	})

	t.Run("should return bad request if diceIndex is not a number", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/lock/fail", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusBadRequest; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})

	t.Run("should return bad request if controller returns error", func(t *testing.T) {
		c := &controllerStub{
			ToggleReturnsError: errors.New("just fail"),
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/lock/2", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusBadRequest; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

func TestGameHandler_roll(t *testing.T) {
	t.Run("should return the dices", func(t *testing.T) {
		want := []*models.Dice{
			&models.Dice{2, false},
		}
		c := &controllerStub{
			RollReturnsDice: want,
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		g.RollCount = 1
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/roll", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
		wantJSON, _ := json.Marshal(map[string]interface{}{
			"Dices":     want,
			"RollCount": 1,
		})
		if got, want := strings.TrimSpace(rr.Body.String()), string(wantJSON); got != want {
			t.Errorf("wrong body: got %q want %q", got, want)
		}
	})

	t.Run("should call roll on controller", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/roll", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := c.RollCallCount, 1; got != want {
			t.Fatalf("wrong number of calls: got %v want %v", got, want)
		}
		if got, want := c.RollCallParamG[0], g; got != want {
			t.Fatalf("wrong g param for call: got %v want %v", got, want)
		}
		if got, want := c.RollCallParamPlayer[0], "alice"; got != want {
			t.Fatalf("wrong player param for call: got %v want %v", got, want)
		}
	})

	t.Run("should return bad request if controller returns error", func(t *testing.T) {
		c := &controllerStub{
			RollReturnsError: errors.New("just fail"),
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/roll", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusBadRequest; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

func TestGameHandler_score(t *testing.T) {
	t.Run("should return 200", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/score", strings.NewReader("yahtzee"))
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})

	t.Run("should call scrore on controller", func(t *testing.T) {
		c := &controllerStub{}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/score", strings.NewReader("yahtzee"))
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := c.ScoreCallCount, 1; got != want {
			t.Fatalf("wrong number of calls: got %v want %v", got, want)
		}
		if got, want := c.ScoreCallParamG[0], g; got != want {
			t.Fatalf("wrong g param for call: got %v want %v", got, want)
		}
		if got, want := c.ScoreCallParamPlayer[0], "alice"; got != want {
			t.Fatalf("wrong player param for call: got %v want %v", got, want)
		}
		if got, want := c.ScoreCallParamC[0], models.Category("yahtzee"); got != want {
			t.Fatalf("wrong c param for call: got %v want %v", got, want)
		}
	})

	t.Run("should return bad request if controller returns error", func(t *testing.T) {
		c := &controllerStub{
			ScoreReturns: errors.New("just fail"),
		}
		h := &GameHandler{
			Controller: c,
		}
		g := models.NewGame()
		g.Players = append(g.Players, models.NewPlayer("alice"))
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/score", strings.NewReader("yahtzee"))
		if err != nil {
			t.Fatal(err)
		}

		h.handle("alice", g).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusBadRequest; got != want {
			t.Errorf("wrong status code: got %v want %v", got, want)
		}
	})
}

type controllerStub struct {
	AddPlayerCallCount     int
	AddPlayerCallParamG    []*models.Game
	AddPlayerCallParamName []string
	AddPlayerReturns       error

	RollCallCount       int
	RollCallParamG      []*models.Game
	RollCallParamPlayer []string
	RollReturnsDice     []*models.Dice
	RollReturnsError    error

	ToggleCallCount          int
	ToggleCallParamG         []*models.Game
	ToggleCallParamPlayer    []string
	ToggleCallParamDiceIndex []int
	ToggleReturnsDice        []*models.Dice
	ToggleReturnsError       error

	ScoreCallCount       int
	ScoreCallParamG      []*models.Game
	ScoreCallParamPlayer []string
	ScoreCallParamC      []models.Category
	ScoreReturns         error
}

func (s *controllerStub) AddPlayer(g *models.Game, name string) error {
	s.AddPlayerCallCount++
	s.AddPlayerCallParamG = append(s.AddPlayerCallParamG, g)
	s.AddPlayerCallParamName = append(s.AddPlayerCallParamName, name)

	return s.AddPlayerReturns
}

func (s *controllerStub) Roll(g *models.Game, player string) ([]*models.Dice, error) {
	s.RollCallCount++
	s.RollCallParamG = append(s.RollCallParamG, g)
	s.RollCallParamPlayer = append(s.RollCallParamPlayer, player)

	return s.RollReturnsDice, s.RollReturnsError
}

func (s *controllerStub) Toggle(g *models.Game, player string, diceIndex int) ([]*models.Dice, error) {
	s.ToggleCallCount++
	s.ToggleCallParamG = append(s.ToggleCallParamG, g)
	s.ToggleCallParamPlayer = append(s.ToggleCallParamPlayer, player)
	s.ToggleCallParamDiceIndex = append(s.ToggleCallParamDiceIndex, diceIndex)

	return s.ToggleReturnsDice, s.ToggleReturnsError
}

func (s *controllerStub) Score(g *models.Game, player string, c models.Category) error {
	s.ScoreCallCount++
	s.ScoreCallParamG = append(s.ScoreCallParamG, g)
	s.ScoreCallParamPlayer = append(s.ScoreCallParamPlayer, player)
	s.ScoreCallParamC = append(s.ScoreCallParamC, c)

	return s.ScoreReturns
}
