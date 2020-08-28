package handler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/pkg/models"
	"github.com/akarasz/yahtzee/pkg/store"
)

type RootHandler struct {
	store store.Store
	game  gameHandler
}

func New(store store.Store, gameHandler *GameHandler) *RootHandler {
	return &RootHandler{
		store: store,
		game:  gameHandler,
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logrus.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})
	ctx := context.WithValue(r.Context(), "logger", log)
	h.serve(w, r.WithContext(ctx))
}

func (h *RootHandler) serve(w http.ResponseWriter, r *http.Request) {
	log := logFrom(r.Context())

	log.Info("incoming request")

	user, _, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "use basic auth for setting your name", http.StatusUnauthorized)
		return
	}

	var id string

	id, r.URL.Path = shiftPath(r.URL.Path)
	switch id {
	case "":
		h.create(w, r)
	default:
		h.load(user, id).ServeHTTP(w, r)
	}
}

func (h *RootHandler) create(w http.ResponseWriter, r *http.Request) {
	log := logFrom(r.Context())

	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	id := generateID()

	err := h.store.Put(id, models.NewGame())
	if err != nil {
		http.Error(w, "unable to create game", http.StatusInternalServerError)
		return
	}

	log.WithField("gameID", id).Info("game created")

	w.Header().Set("Location", fmt.Sprintf("/%s", id))
	w.WriteHeader(http.StatusCreated)
}

func (h *RootHandler) load(user string, id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g, err := h.store.Get(id)
		if err != nil {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}

		h.game.handle(user, g).ServeHTTP(w, r.WithContext(
			logWithFields(r.Context(), logrus.Fields{
				"gameID": id,
				"user":   user,
			})))
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
