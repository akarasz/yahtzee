package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/pkg/game"
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
	finalHandler := http.HandlerFunc(h.serve)
	allowCors(contextLogger(finalHandler)).ServeHTTP(w, r)
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
	case "score":
		h.score(w, r)
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

func (h *RootHandler) score(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	if numberOfDices := len(r.URL.Query()["dice"]); numberOfDices != 5 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	dices := make([]int, 5)
	for i, d := range r.URL.Query()["dice"] {
		v, err := strconv.Atoi(d)
		if err != nil || v < 1 || 6 < v {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		dices[i] = v
	}

	categories := []models.Category{
		models.Ones,
		models.Twos,
		models.Threes,
		models.Fours,
		models.Fives,
		models.Sixes,
		models.ThreeOfAKind,
		models.FourOfAKind,
		models.FullHouse,
		models.SmallStraight,
		models.LargeStraight,
		models.Yahtzee,
		models.Chance,
	}

	result := map[models.Category]int{}
	for _, c := range categories {
		score, err := game.Score(c, dices)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		result[c] = score
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
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
		length    = 4
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}
