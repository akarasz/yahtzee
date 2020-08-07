package handler

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/store"
)

type RootHandler struct {
	store *store.Store
	game  *GameHandler
}

func New(store *store.Store) *RootHandler {
	return &RootHandler{
		store: store,
		game:  &GameHandler{},
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string

	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "":
		h.create(w, r)
	default:
		h.id(head).ServeHTTP(w, r)
	}
}

func (h *RootHandler) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	id := generateID()

	err := h.store.Put(id, game.New())
	if err != nil {
		http.Error(w, "unable to create game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", id))
	w.WriteHeader(http.StatusCreated)
}

func (h *RootHandler) id(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g, err := h.store.Get(id)
		if err != nil {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}

		h.game.handle(g).ServeHTTP(w, r)
	})
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
