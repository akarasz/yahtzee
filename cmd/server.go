package main

import (
	"net/http"

	"github.com/akarasz/yahtzee/pkg/handler"
)

func main() {
	h := handler.New()

	http.ListenAndServe(":8000", h)
}
