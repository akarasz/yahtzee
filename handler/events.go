package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/akarasz/yahtzee/events"
	"github.com/akarasz/yahtzee/store"
)

const (
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 8) / 10
)

var upgrader = websocket.Upgrader{}

func writer(ws *websocket.Conn, events <-chan interface{}) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
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

func reader(ws *websocket.Conn) {
	defer func() {
		ws.Close()
	}()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func EventsWSHandler(sub events.Subscriber, s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameID := mux.Vars(r)["gameID"]
		if _, err := s.Load(gameID); err != nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		eventChannel, err := sub.Subscribe(gameID)
		if err != nil {
			http.Error(w, "Unable to subscribe", http.StatusInternalServerError)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				http.Error(w, "Unknown error", http.StatusInternalServerError)
			}
			return
		}

		go writer(ws, eventChannel)
		reader(ws)
	})
}
