package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/akarasz/yahtzee/controller"
	"github.com/akarasz/yahtzee/events"
	"github.com/akarasz/yahtzee/handler"
	"github.com/akarasz/yahtzee/service"
	"github.com/akarasz/yahtzee/store"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	rand.Seed(time.Now().UnixNano())

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS"),
	})
	defer rdb.Close()

	e := events.New()
	sp := service.NewProvider()
	s := store.NewRedis(rdb, 30*time.Minute)

	c := controller.New(s, sp, e, redislock.New(rdb))
	h := handler.New(c, c)

	r := mux.NewRouter()
	r.Use(
		handler.CorsMiddleware,
		handler.ContextLoggerMiddleware)
	r.HandleFunc("/", h.CreateHandler).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/score", h.ScoresHandler).
		Methods("GET", "OPTIONS").
		Queries("dices", "{dices:[1-6],[1-6],[1-6],[1-6],[1-6]}")
	r.HandleFunc("/{gameID}", h.GetHandler).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{gameID}/join", h.AddPlayerHandler).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/roll", h.RollHandler).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/lock/{dice}", h.LockHandler).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{gameID}/score", h.ScoreHandler).
		Methods("POST", "OPTIONS")
	r.Handle("/{gameID}/ws", handler.EventsWSHandler(e, s))

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
