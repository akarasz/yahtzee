package main

import (
	"net/http"

	"github.com/akarasz/yahtzee/pkg/handler"
)

func main() {
	h := handler.New()

	err := http.ListenAndServe(":8000", h)
	if err != nil {
		panic(err)
	}
}
