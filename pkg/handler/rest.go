package handler

import (
	"context"
	"fmt"
	"io/ioutil"
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

type RootHandler struct {
	game *GameHandler
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string

	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		switch r.Method {
		case "POST":
			fmt.Fprint(w, "create new game")
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	default:
		h.game.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "gameId", head)))
	}
}

func New() *RootHandler {
	return &RootHandler{
		game: &GameHandler{},
	}
}

type GameHandler struct {
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("gameId").(string)

	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)

	switch head {
	case "":
		switch r.Method {
		case "GET":
			fmt.Fprintf(w, "getting game %q", id)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	case "join":
		switch r.Method {
		case "POST":
			fmt.Fprintf(w, "join game %q", id)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	case "lock":
		switch r.Method {
		case "POST":
			head, r.URL.Path = shiftPath(r.URL.Path)

			dice, err := strconv.Atoi(head)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			fmt.Fprintf(w, "lock dice %d in game %q", dice, id)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	case "roll":
		switch r.Method {
		case "POST":
			fmt.Fprintf(w, "roll in game %q", id)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	case "score":
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			bodyString := string(body)
			fmt.Fprintf(w, "score in game %q, category is %q", id, bodyString)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}
