package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/akarasz/yahtzee/controller"
	"github.com/akarasz/yahtzee/models"
	"github.com/akarasz/yahtzee/store"
)

type Root interface {
	CreateHandler(w http.ResponseWriter, r *http.Request)
	GetHandler(w http.ResponseWriter, r *http.Request)
	ScoresHandler(w http.ResponseWriter, r *http.Request)
}

type Game interface {
	AddPlayerHandler(w http.ResponseWriter, r *http.Request)
	RollHandler(w http.ResponseWriter, r *http.Request)
	LockHandler(w http.ResponseWriter, r *http.Request)
	ScoreHandler(w http.ResponseWriter, r *http.Request)
}

type Default struct {
	rootController controller.Root
	gameController controller.Game
}

func New(root controller.Root, game controller.Game) *Default {
	return &Default{
		rootController: root,
		gameController: game,
	}
}

func (h *Default) CreateHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := h.rootController.Create()
	if err != nil {
		log.Print(err)
		http.Error(w, "unable to create game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", gameID))
	w.WriteHeader(http.StatusCreated)
}

func (h *Default) GetHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := extractGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	g, err := h.rootController.Get(gameID)

	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(g); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Default) ScoresHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.rootController.Scores(strings.Split(mux.Vars(r)["dices"], ","))
	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Default) AddPlayerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.gameController.AddPlayer(user, gameID)

	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Default) RollHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.gameController.Roll(user, gameID)

	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Default) LockHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dice, ok := mux.Vars(r)["dice"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.gameController.Lock(user, gameID, dice)

	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Default) ScoreHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	bodyString := string(body)

	res, err := h.gameController.Score(user, gameID, models.Category(bodyString))

	if controllerHasError(err, w) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func controllerHasError(err error, w http.ResponseWriter) bool {
	if errors.As(err, &store.ErrNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return true
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}

	return false
}

func extractGameID(r *http.Request) (string, error) {
	gameID, ok := mux.Vars(r)["gameID"]
	if !ok {
		return "", errors.New("no gameID")
	}
	return gameID, nil
}

func extractUser(r *http.Request) (*models.User, error) {
	var res models.User
	user, _, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("no user")
	}
	res = models.User(user)
	return &res, nil
}
