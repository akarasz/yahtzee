package handler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
)

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
