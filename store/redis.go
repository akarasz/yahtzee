package store

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/akarasz/yahtzee/models"
)

var ctx = context.Background()

type Redis struct {
	c *redis.Client
}

func (r *Redis) Load(id string) (models.Game, error) {
	return models.Game{}, nil
}

func (r *Redis) Save(id string, g models.Game) error {
	return nil
}
