package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/handler"
	"github.com/akarasz/yahtzee/pkg/store"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		ForceQuote:       true,
	})

	fmt.Println("[info] message key=value level=info message=message")
	fmt.Println("[error] message key=value level=info message=message")

	rand.Seed(time.Now().UnixNano())

	h := handler.New(
		store.NewInMemory(),
		&handler.GameHandler{
			Controller: game.New(),
		})

	port := "8000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	listenAddress := ":" + port

	log.Infoln("starting server on", listenAddress)
	err := http.ListenAndServe(listenAddress, h)
	if err != nil {
		log.Errorln("listen and serve:", err)
		panic(err)
	}
}
