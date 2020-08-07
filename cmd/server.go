package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/akarasz/yahtzee/pkg/handler"
	"github.com/akarasz/yahtzee/pkg/store"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	s := store.NewInMemory()
	h := handler.New(s)

	err := http.ListenAndServe(":8000", h)
	if err != nil {
		panic(err)
	}
}
