package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Location")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS"),
	})

	e := events.New()
	sp := service.NewProvider()
	s := store.NewRedis(rdb, 30*time.Minute)

	c := controller.New(s, sp, e)
	h := handler.New(c, c)

	r := mux.NewRouter()
	r.Use(corsMiddleware)
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
