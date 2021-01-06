package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/akarasz/yahtzee/model"
	"github.com/akarasz/yahtzee/store"
)

var ctx = context.Background()

type Redis struct {
	client     *redis.Client
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
		expiration: expiration,
	}
}

func (r *Redis) Load(id string) (model.Game, error) {
	var res model.Game

	raw, err := r.client.Get(ctx, "game:"+id).Bytes()
	if err != nil {
		return model.Game{}, err
	}

	err = json.Unmarshal(raw, &res)

	return res, err
}

func (r *Redis) Save(id string, g model.Game) error {
	raw, err := json.Marshal(g)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "game:"+id, string(raw), r.expiration).Err()
}