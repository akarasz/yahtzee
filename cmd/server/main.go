package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"

	event "github.com/akarasz/yahtzee/event/rabbit"
	"github.com/akarasz/yahtzee/handler"
	store "github.com/akarasz/yahtzee/store/redis"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// redis
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS"),
	})
	defer rdb.Close()
	s := store.New(rdb, 48*time.Hour)

	// rabbit
	rabbitConn, err := amqp.Dial(os.Getenv("RABBIT"))
	if err != nil {
		panic(err)
	}
	defer rabbitConn.Close()
	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		panic(err)
	}
	defer rabbitChan.Close()
	e, err := event.New(rabbitChan)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	port := "8000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	listenAddress := ":" + port
	log.Fatal(http.ListenAndServe(listenAddress, handler.New(s, e, e)))
}
