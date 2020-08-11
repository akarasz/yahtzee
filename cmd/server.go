package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/handler"
	"github.com/akarasz/yahtzee/pkg/store"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	h := handler.New(
		store.NewInMemory(),
		&handler.GameHandler{
			game.New(),
		})

	err := http.ListenAndServe(":8000", h)
	if err != nil {
		panic(err)
	}
}
