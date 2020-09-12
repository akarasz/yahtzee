package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/controller"
	"github.com/akarasz/yahtzee/handler"
	"github.com/akarasz/yahtzee/service"
	"github.com/akarasz/yahtzee/store"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	sp := service.NewProvider()
	s := store.New()
	c := controller.New(s, sp)
	h := handler.New(c, c)

	r := mux.NewRouter()
	r.HandleFunc("/", h.CreateHandler).
		Methods("POST")
	r.HandleFunc("/score", h.ScoresHandler).
		Methods("GET").
		Queries("dices", "{dices:[1-6],[1-6],[1-6],[1-6],[1-6]}")
	r.HandleFunc("/{gameID}", h.GetHandler).
		Methods("GET")
	r.HandleFunc("/{gameID}/join", h.AddPlayerHandler).
		Methods("POST")
	r.HandleFunc("/{gameID}/roll", h.RollHandler).
		Methods("POST")
	r.HandleFunc("/{gameID}/lock/{dice}", h.LockHandler).
		Methods("POST")
	r.HandleFunc("/{gameID}/score", h.ScoreHandler).
		Methods("POST")

	port := "8000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	listenAddress := ":" + port
	log.Fatal(http.ListenAndServe(listenAddress, r))
}
