package main

import (
	"net/http"

	"github.com/akarasz/yahtzee/pkg/handler"
	"github.com/akarasz/yahtzee/pkg/store"
)

func main() {
	s := store.New()
	h := handler.New(s)

	err := http.ListenAndServe(":8000", h)
	if err != nil {
		panic(err)
	}
}
