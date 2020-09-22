package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/store"
)

var upgrader = websocket.Upgrader{}

func EventsWSHandler(s store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameID := mux.Vars(r)["gameID"]
		if _, err := s.Load(gameID); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.WithField("gameID", gameID).Printf("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("gameID", gameID).Printf("read:", err)
				break
			}
			log.WithField("gameID", gameID).Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.WithField("gameID", gameID).Println("write:", err)
				break
			}
		}
	})
}
