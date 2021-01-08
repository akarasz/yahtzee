package yahtzee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/model"
	"github.com/akarasz/yahtzee/store"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type handler struct {
	store      store.Store
	emitter    event.Emitter
	subscriber event.Subscriber
}

func NewHandler(s store.Store, e event.Emitter, sub event.Subscriber) http.Handler {
	h := &handler{s, e, sub}

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.HandleFunc("/", h.Create).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/score", h.Hints).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{gameID}", h.Get).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{gameID}/join", h.AddPlayer).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/roll", h.Roll).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/lock/{dice}", h.Lock).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/score", h.Score).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/ws", h.WS)
	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Location")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			return
		}

		next.ServeHTTP(w, r)
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

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	gameID := generateID()
	if err := h.store.Save(gameID, *model.NewGame()); err != nil {
		writeError(w, r, err, "create game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", gameID))
	w.WriteHeader(http.StatusCreated)

	log.Print("game created")
}

func (h *handler) Hints(w http.ResponseWriter, r *http.Request) {
	dices, ok := readDices(w, r)
	if !ok {
		return
	}

	res := map[model.Category]int{}
	for _, c := range model.Categories() {
		score, err := score(c, dices)
		if err != nil {
			writeError(w, r, err, "", http.StatusInternalServerError)
			return
		}
		res[c] = score
	}

	if ok := writeJSON(w, r, res); !ok {
		return
	}

	log.Print("hints returned")
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}

	g, err := h.store.Load(gameID)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	if ok := writeJSON(w, r, g); !ok {
		return
	}

	log.Print("game returned")
}

type AddPlayerResponse struct {
	Players []*model.Player
}

func (h *handler) AddPlayer(w http.ResponseWriter, r *http.Request) {
	user, ok := readUser(w, r)
	if !ok {
		return
	}
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}

	g, err := h.store.Load(gameID)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	if g.CurrentPlayer > 0 || g.Round > 0 {
		writeError(w, r, nil, "game already started", http.StatusBadRequest)
		return
	}
	for _, p := range g.Players {
		if p.User == user {
			writeError(w, r, nil, "already joined", http.StatusConflict)
			return
		}
	}

	g.Players = append(g.Players, model.NewPlayer(user))

	if err := h.store.Save(gameID, g); err != nil {
		writeStoreError(w, r, err)
		return
	}

	changes := &AddPlayerResponse{
		Players: g.Players,
	}

	h.emitter.Emit(gameID, &user, event.AddPlayer, changes)

	w.WriteHeader(http.StatusCreated)
	if ok := writeJSON(w, r, changes); !ok {
		return
	}

	log.Print("player added")
}

type RollResponse struct {
	Dices     []*model.Dice
	RollCount int
}

func (h *handler) Roll(w http.ResponseWriter, r *http.Request) {
	user, ok := readUser(w, r)
	if !ok {
		return
	}
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}

	g, err := h.store.Load(gameID)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	if len(g.Players) == 0 {
		writeError(w, r, nil, "no players joined", http.StatusBadRequest)
		return
	}
	currentPlayer := g.Players[g.CurrentPlayer]
	if user != currentPlayer.User {
		writeError(w, r, nil, "another players turn", http.StatusBadRequest)
		return
	}
	if g.Round >= 13 {
		writeError(w, r, nil, "game is over", http.StatusBadRequest)
		return
	}
	if g.RollCount >= 3 {
		writeError(w, r, nil, "no more rolls", http.StatusBadRequest)
		return
	}

	for _, d := range g.Dices {
		if d.Locked {
			continue
		}

		d.Value = rand.Intn(6) + 1
	}

	g.RollCount++

	if err := h.store.Save(gameID, g); err != nil {
		writeStoreError(w, r, err)
		return
	}

	changes := &RollResponse{
		Dices:     g.Dices,
		RollCount: g.RollCount,
	}

	h.emitter.Emit(gameID, &user, event.Roll, changes)

	if ok := writeJSON(w, r, changes); !ok {
		return
	}

	log.Print("rolled dices")
}

type LockResponse struct {
	Dices []*model.Dice
}

func (h *handler) Lock(w http.ResponseWriter, r *http.Request) {
	user, ok := readUser(w, r)
	if !ok {
		return
	}
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}
	diceIndex, ok := readDiceIndex(w, r)
	if !ok {
		return
	}

	g, err := h.store.Load(gameID)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	if len(g.Players) == 0 {
		writeError(w, r, nil, "no players joined", http.StatusBadRequest)
		return
	}
	currentPlayer := g.Players[g.CurrentPlayer]
	if user != currentPlayer.User {
		writeError(w, r, nil, "another players turn", http.StatusBadRequest)
		return
	}
	if g.Round >= 13 {
		writeError(w, r, nil, "game is over", http.StatusBadRequest)
		return
	}
	if g.RollCount == 0 {
		writeError(w, r, nil, "roll first", http.StatusBadRequest)
		return
	}
	if g.RollCount >= 3 {
		writeError(w, r, nil, "no more rolls", http.StatusBadRequest)
		return
	}

	g.Dices[diceIndex].Locked = !g.Dices[diceIndex].Locked

	if err := h.store.Save(gameID, g); err != nil {
		writeStoreError(w, r, err)
		return
	}

	changes := &LockResponse{
		Dices: g.Dices,
	}

	h.emitter.Emit(gameID, &user, event.Lock, changes)

	if ok := writeJSON(w, r, changes); !ok {
		return
	}

	log.Print("toggled dice")
}

func (h *handler) Score(w http.ResponseWriter, r *http.Request) {
	user, ok := readUser(w, r)
	if !ok {
		return
	}
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}
	category, ok := readCategory(w, r)
	if !ok {
		return
	}

	g, err := h.store.Load(gameID)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}

	if len(g.Players) == 0 {
		writeError(w, r, nil, "no players joined", http.StatusBadRequest)
		return
	}
	currentPlayer := g.Players[g.CurrentPlayer]
	if user != currentPlayer.User {
		writeError(w, r, nil, "another players turn", http.StatusBadRequest)
		return
	}
	if g.Round >= 13 {
		writeError(w, r, nil, "game is over", http.StatusBadRequest)
		return
	}
	if g.RollCount == 0 {
		writeError(w, r, nil, "roll first", http.StatusBadRequest)
		return
	}
	if _, ok := currentPlayer.ScoreSheet[category]; ok {
		writeError(w, r, nil, "category is already used", http.StatusBadRequest)
		return
	}

	dices := make([]int, 5)
	for i, d := range g.Dices {
		dices[i] = d.Value
	}

	score, err := score(category, dices)
	if err != nil {
		writeError(w, r, err, "invalid category", http.StatusBadRequest)
		return
	}

	currentPlayer.ScoreSheet[category] = score

	if _, ok := currentPlayer.ScoreSheet[model.Bonus]; !ok {
		var total, types int
		for k, v := range currentPlayer.ScoreSheet {
			if k == model.Ones || k == model.Twos || k == model.Threes ||
				k == model.Fours || k == model.Fives || k == model.Sixes {
				types++
				total += v
			}
		}

		if total >= 63 {
			currentPlayer.ScoreSheet[model.Bonus] = 35
		} else if types == 6 {
			currentPlayer.ScoreSheet[model.Bonus] = 0
		}
	}

	for _, d := range g.Dices {
		d.Locked = false
	}

	g.RollCount = 0
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
	if g.CurrentPlayer == 0 {
		g.Round++
	}

	if err := h.store.Save(gameID, g); err != nil {
		writeStoreError(w, r, err)
		return
	}

	h.emitter.Emit(gameID, &user, event.Score, &g)

	if ok := writeJSON(w, r, &g); !ok {
		return
	}

	log.Print("scored")
}

const (
	wsPongWait   = 30 * time.Second
	wsPingPeriod = (wsPongWait * 8) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsWriter(ws *websocket.Conn, events <-chan *event.Event, s event.Subscriber, gameID string) {
	pingTicker := time.NewTicker(wsPingPeriod)
	defer func() {
		s.Unsubscribe(gameID, ws)
		pingTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case e := <-events:
			if err := ws.WriteJSON(e); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func wsReader(ws *websocket.Conn, s event.Subscriber, gameID string) {
	defer func() {
		s.Unsubscribe(gameID, ws)
		ws.Close()
	}()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(wsPongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(wsPongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *handler) WS(w http.ResponseWriter, r *http.Request) {
	gameID, ok := readGameID(w, r)
	if !ok {
		return
	}

	if _, err := h.store.Load(gameID); err != nil {
		writeStoreError(w, r, err)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			writeError(w, r, err, "unknown error", http.StatusInternalServerError)
		}
		return
	}

	eventChannel, err := h.subscriber.Subscribe(gameID, ws)
	if err != nil {
		writeError(w, r, err, "unable to subscribe", http.StatusInternalServerError)
		return
	}

	go wsWriter(ws, eventChannel, h.subscriber, gameID)
	wsReader(ws, h.subscriber, gameID)
}

func readDiceIndex(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw, ok := mux.Vars(r)["dice"]
	if !ok {
		writeError(w, r, nil, "no dice index in request", http.StatusInternalServerError)
		return 0, false
	}
	index, err := strconv.Atoi(raw)
	if err != nil || index < 0 || index > 4 {
		writeError(w, r, err, "invalid dice index", http.StatusBadRequest)
		return index, false
	}
	return index, true
}

func readDices(w http.ResponseWriter, r *http.Request) ([]int, bool) {
	raw := r.URL.Query().Get("dices")
	rawDices := strings.Split(raw, ",")
	if len(rawDices) != 5 {
		writeError(w, r, nil, "wrong number of dices", http.StatusBadRequest)
		return nil, false
	}
	dices := make([]int, 5)
	for i, d := range rawDices {
		v, err := strconv.Atoi(d)
		if err != nil || v < 1 || 6 < v {
			writeError(w, r, err, "invalid dice", http.StatusBadRequest)
			return nil, false
		}
		dices[i] = v
	}
	return dices, true
}

func readCategory(w http.ResponseWriter, r *http.Request) (model.Category, bool) {
	if r.Body == nil {
		writeError(w, r, nil, "no category", http.StatusBadRequest)
		return "", false
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, err, "extract category from body", http.StatusInternalServerError)
		return "", false
	}
	return model.Category(body), true
}

func readGameID(w http.ResponseWriter, r *http.Request) (string, bool) {
	gameID, ok := mux.Vars(r)["gameID"]
	if !ok {
		err := errors.New("no gameID")
		writeError(w, r, err, "no gameID in request", http.StatusInternalServerError)
		return "", false
	}
	return gameID, true
}

func readUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	user, _, ok := r.BasicAuth()
	if !ok {
		err := errors.New("no user")
		writeError(w, r, err, "no user in request", http.StatusUnauthorized)
		return "", false
	}
	return model.User(user), true
}

func writeJSON(w http.ResponseWriter, r *http.Request, body interface{}) bool {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		writeError(w, r, err, "response json encode", http.StatusInternalServerError)
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, r *http.Request, err error, msg string, status int) {
	log.Printf("%s: %v", msg, err)
	http.Error(w, "", status)
}

func writeStoreError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.As(err, &store.ErrNotExists) {
		writeError(w, r, err, "not exists", http.StatusNotFound)
	} else {
		writeError(w, r, err, "unknown error", http.StatusInternalServerError)
	}
}

func score(category model.Category, dices []int) (int, error) {
	s := 0
	switch category {
	case model.Ones:
		for _, d := range dices {
			if d == 1 {
				s++
			}
		}
	case model.Twos:
		for _, d := range dices {
			if d == 2 {
				s += 2
			}
		}
	case model.Threes:
		for _, d := range dices {
			if d == 3 {
				s += 3
			}
		}
	case model.Fours:
		for _, d := range dices {
			if d == 4 {
				s += 4
			}
		}
	case model.Fives:
		for _, d := range dices {
			if d == 5 {
				s += 5
			}
		}
	case model.Sixes:
		for _, d := range dices {
			if d == 6 {
				s += 6
			}
		}
	case model.ThreeOfAKind:
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}

		for k, v := range occurrences {
			if v >= 3 {
				s = 3 * k
			}
		}
	case model.FourOfAKind:
		occurrences := map[int]int{}
		for _, d := range dices {
			occurrences[d]++
		}

		for k, v := range occurrences {
			if v >= 4 {
				s = 4 * k
			}
		}
	case model.FullHouse:
		one, oneCount, other := dices[0], 1, 0
		for i := 1; i < len(dices); i++ {
			v := dices[i]

			if one == v {
				oneCount++
			} else if other == 0 || other == v {
				other = v
			} else {
				oneCount = 4
			}
		}

		if oneCount == 2 || oneCount == 3 {
			s = 25
		}
	case model.SmallStraight:
		hit := [6]bool{}
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3]) ||
			(hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 30
		}
	case model.LargeStraight:
		hit := [6]bool{}
		for _, d := range dices {
			hit[d-1] = true
		}

		if (hit[0] && hit[1] && hit[2] && hit[3] && hit[4]) ||
			(hit[1] && hit[2] && hit[3] && hit[4] && hit[5]) {
			s = 40
		}
	case model.Yahtzee:
		same := true
		for i := 0; i < len(dices)-1; i++ {
			same = same && dices[i] == dices[i+1]
		}

		if same {
			s = 50
		}
	case model.Chance:
		for _, d := range dices {
			s += d
		}
	default:
		return 0, errors.New("invalid category")
	}

	return s, nil
}
