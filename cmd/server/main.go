package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/akarasz/yahtzee"
	event "github.com/akarasz/yahtzee/event/rabbit"
	store "github.com/akarasz/yahtzee/store/redis"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS"),
	})
	defer rdb.Close()
	s := store.New(rdb, 48*time.Hour)

	e, err := event.New(os.Getenv("RABBIT"))
	if err != nil {
		panic(err)
	}
	defer e.Close()

	r := yahtzee.NewHandler(s, e, e)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	port := "8000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	listenAddress := ":" + port
	log.Fatal(http.ListenAndServe(listenAddress, r))
}
