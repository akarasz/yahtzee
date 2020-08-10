package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/models"
)

type gameHandler interface {
	handle(player string, game *models.Game) http.Handler
}

type GameHandler struct{}

func (h *GameHandler) handle(player string, g *models.Game) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var head string
		head, r.URL.Path = shiftPath(r.URL.Path)

		c := game.New(player, g)
		switch head {
		case "":
			h.root(c).ServeHTTP(w, r)
		case "join":
			h.join(c).ServeHTTP(w, r)
		case "lock":
			h.lock(c).ServeHTTP(w, r)
		case "roll":
			h.roll(c).ServeHTTP(w, r)
		case "score":
			h.score(c).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func (h *GameHandler) root(c game.Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(c.Snapshot()); err != nil {
			panic(err)
		}
	})
}

func (h *GameHandler) join(c game.Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		if err := c.AddPlayer(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func (h *GameHandler) lock(c game.Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		var head string
		head, r.URL.Path = shiftPath(r.URL.Path)

		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		dice, err := strconv.Atoi(head)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		res, err := c.Toggle(dice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			panic(err)
		}
	})
}

func (h *GameHandler) roll(c game.Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		res, err := c.Roll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			panic(err)
		}
	})
}

func (h *GameHandler) score(c game.Controller) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		bodyString := string(body)

		if err := c.Score(models.Category(bodyString)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
