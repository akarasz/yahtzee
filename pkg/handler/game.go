package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/akarasz/yahtzee/pkg/game"
)

type GameHandler struct {
	id string
}

func (h *GameHandler) handle(g *game.Game, user string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var head string
		head, r.URL.Path = shiftPath(r.URL.Path)

		switch head {
		case "":
			h.root(g).ServeHTTP(w, r)
		case "join":
			h.join(g, user).ServeHTTP(w, r)
		case "lock":
			h.lock(g).ServeHTTP(w, r)
		case "roll":
			h.roll(g).ServeHTTP(w, r)
		case "score":
			h.score(g).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func (h *GameHandler) root(g *game.Game) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewEncoder(w).Encode(g); err != nil {
			panic(err)
		}
	})
}

func (h *GameHandler) join(g *game.Game, user string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		g.AddPlayer(user)

		w.WriteHeader(http.StatusCreated)
	})
}

func (h *GameHandler) lock(g *game.Game) http.Handler {
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

		fmt.Fprintf(w, "lock dice %d", dice)
	})
}

func (h *GameHandler) roll(g *game.Game) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprint(w, "roll in game")
	})
}

func (h *GameHandler) score(g *game.Game) http.Handler {
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
		fmt.Fprintf(w, "score category %q", bodyString)
	})
}
