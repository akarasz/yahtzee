package main

import (
	"net/http"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/handler"
	"github.com/akarasz/yahtzee/pkg/store"
)

func main() {
	repo := map[string]*game.Game{}
	s := store.New(repo)
	h := handler.New(s)

	err := http.ListenAndServe(":8000", h)
	if err != nil {
		panic(err)
	}
}
