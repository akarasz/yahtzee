package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/store"
)

var ctx = context.Background()

var (
	lockExpiration = 5 * time.Second
	lockBackoff    = redislock.LinearBackoff(50 * time.Millisecond)
)

type Redis struct {
	client     *redis.Client
	locker     *redislock.Client
	expiration time.Duration
}

func New(client *redis.Client, expiration time.Duration) store.Store {
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "yahtzee_redis_store_size",
			Help: "The total number of games in the redis store",
		},
		func() float64 {
			return float64(client.DBSize(ctx).Val())
		})

	return &Redis{
		client:     client,
		locker:     redislock.New(client),
		expiration: expiration,
	}
}

func (r *Redis) Load(id string) (yahtzee.Game, error) {
	var res yahtzee.Game

	raw, err := r.client.Get(ctx, "game:"+id).Bytes()
	if err != nil {
		return yahtzee.Game{}, store.ErrNotExists
	}

	err = json.Unmarshal(raw, &res)

	res.Scorer = yahtzee.ComposeScorer(res.Features...)

	return res, err
}

func (r *Redis) Save(id string, g yahtzee.Game) error {
	raw, err := json.Marshal(g)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "game:"+id, string(raw), r.expiration).Err()
}

func (r *Redis) Lock(id string) (func(), error) {
	lock, err := r.locker.Obtain(
		context.Background(),
		"lock:"+id,
		lockExpiration,
		&redislock.Options{
			RetryStrategy: lockBackoff,
		})

	if err != nil {
		if err == redislock.ErrNotObtained {
			log.Println("could not obtain lock")
		} else if err != nil {
			log.Fatalln(err)
		}
		return nil, err
	}

	return func() { lock.Release(context.Background()) }, nil
}
