package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/akarasz/yahtzee/controller"
	"github.com/akarasz/yahtzee/model"
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
		handleError(w, r, err, "create game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", gameID))
	w.WriteHeader(http.StatusCreated)

	LogFrom(r).Info("game created")
}

func (h *Default) GetHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := extractGameID(r)
	if err != nil {
		handleError(w, r, err, "extract gameID", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	g, err := h.rootController.Get(gameID)
	if handleControllerError(w, r, err, "controller") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(g); err != nil {
		handleError(w, r, err, "response json encode", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Info("game returned")
}

func (h *Default) ScoresHandler(w http.ResponseWriter, r *http.Request) {
	res, err := h.rootController.Scores(strings.Split(mux.Vars(r)["dices"], ","))
	if handleControllerError(w, r, err, "scores") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		handleError(w, r, err, "response json encoding", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Info("scores returned")
}

func (h *Default) AddPlayerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		handleError(w, r, err, "extract user", http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		handleError(w, r, err, "extract gameID", http.StatusBadRequest)
		return
	}

	res, err := h.gameController.AddPlayer(user, gameID)
	if handleControllerError(w, r, err, "add player") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		handleError(w, r, err, "response json encoding", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Info("player added")
}

func (h *Default) RollHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		handleError(w, r, err, "extract user", http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		handleError(w, r, err, "extract gameID", http.StatusBadRequest)
		return
	}

	res, err := h.gameController.Roll(user, gameID)

	if handleControllerError(w, r, err, "roll") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		handleError(w, r, err, "response json encoding", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Info("rolled dices")
}

func (h *Default) LockHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		handleError(w, r, err, "extract user", http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		handleError(w, r, err, "extract gameID", http.StatusBadRequest)
		return
	}

	dice, ok := mux.Vars(r)["dice"]
	AddLogField(r, "dice", dice)
	if !ok {
		handleError(w, r, err, "extract dice index", http.StatusBadRequest)
		return
	}

	res, err := h.gameController.Lock(user, gameID, dice)
	if handleControllerError(w, r, err, "lock") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		handleError(w, r, err, "response json encoding", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Infof("toggled dice")
}

func (h *Default) ScoreHandler(w http.ResponseWriter, r *http.Request) {
	user, err := extractUser(r)
	if err != nil {
		handleError(w, r, err, "extract user", http.StatusUnauthorized)
		return
	}

	gameID, err := extractGameID(r)
	if err != nil {
		handleError(w, r, err, "extract gameID", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError(w, r, err, "extract category from body", http.StatusInternalServerError)
		return
	}
	bodyString := string(body)
	AddLogField(r, "category", bodyString)

	res, err := h.gameController.Score(user, gameID, model.Category(bodyString))
	if handleControllerError(w, r, err, "score") {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		handleError(w, r, err, "response json encoding", http.StatusInternalServerError)
		return
	}

	LogFrom(r).Info("scored")
}

func handleError(w http.ResponseWriter, r *http.Request, err error, msg string, status int) {
	log := LogFrom(r)
	log.Errorf("%s: %v", msg, err)
	http.Error(w, "", status)
}

func handleControllerError(w http.ResponseWriter, r *http.Request, err error, msg string) bool {
	if errors.As(err, &store.ErrNotExists) {
		handleError(w, r, err, msg, http.StatusNotFound)
		return true
	}

	if err != nil {
		handleError(w, r, err, msg, http.StatusInternalServerError)
		return true
	}

	return false
}

func extractGameID(r *http.Request) (string, error) {
	gameID, ok := mux.Vars(r)["gameID"]
	if !ok {
		return "", errors.New("no gameID")
	}
	AddLogField(r, "gid", gameID)
	return gameID, nil
}

func extractUser(r *http.Request) (*model.User, error) {
	var res model.User
	user, _, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("no user")
	}
	res = model.User(user)
	AddLogField(r, "user", res)
	return &res, nil
}
