package main

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/controller"
	"github.com/akarasz/yahtzee/handler"
	"github.com/akarasz/yahtzee/service"
	"github.com/akarasz/yahtzee/store"
)

func main() {
	sp := service.NewProvider()
	s := store.New()
	c := controller.New(s, sp)
	h := handler.New(c, c)

	r := mux.NewRouter()
	r.HandleFunc("/", h.CreateHandler).
		Methods("POST")
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

	log.Fatal(http.ListenAndServe(":8000", r))
}
