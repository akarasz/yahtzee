package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type contextKey string

const gameID contextKey = "gameID"

type RootHandler struct {
	game *GameHandler
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string

	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		h.create(w, r)
	default:
		id := head
		ctx := context.WithValue(r.Context(), gameID, id)
		h.game.ServeHTTP(w, r.WithContext(ctx))
	}
}

func generateID() string {
	const (
		idCharset = "abcdefghijklmnopqrstvwxyz0123456789"
		length    = 5
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}

func (h *RootHandler) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, "create new game %q", generateID())
}

func New() *RootHandler {
	return &RootHandler{
		game: &GameHandler{},
	}
}

type GameHandler struct {
	id string
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(gameID).(string)

	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)

	switch head {
	case "":
		h.root(id).ServeHTTP(w, r)
	case "join":
		h.join(id).ServeHTTP(w, r)
	case "lock":
		h.lock(id).ServeHTTP(w, r)
	case "roll":
		h.roll(id).ServeHTTP(w, r)
	case "score":
		h.score(id).ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *GameHandler) root(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "getting game %q", id)
	})
}

func (h *GameHandler) join(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "join game %q", id)
	})
}

func (h *GameHandler) lock(id string) http.Handler {
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

		fmt.Fprintf(w, "lock dice %d in game %q", dice, id)
	})
}

func (h *GameHandler) roll(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "roll in game %q", id)
	})
}

func (h *GameHandler) score(id string) http.Handler {
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
		fmt.Fprintf(w, "score in game %q, category is %q", h.id, bodyString)
	})
}
