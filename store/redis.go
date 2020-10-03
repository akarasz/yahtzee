package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/akarasz/yahtzee/models"
)

var ctx = context.Background()

type Redis struct {
	client     *redis.Client
	expiration time.Duration
}

func (r *Redis) Load(id string) (models.Game, error) {
	var res models.Game

	raw, err := r.client.Get(ctx, id).Bytes()
	if err != nil {
		return models.Game{}, err
	}

	err = json.Unmarshal(raw, &res)

	return res, err
}

func (r *Redis) Save(id string, g models.Game) error {
	return r.client.Set(ctx, id, g, r.expiration).Err()
}
