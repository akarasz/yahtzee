package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"

	"github.com/akarasz/yahtzee/store"
	redis_store "github.com/akarasz/yahtzee/store/redis"
)

func TestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping store/redis test")
	}

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis",
			ExposedPorts: []string{"6379/tcp"},
		},
		Started: true,
	})
	if err != nil {
		t.Error(err)
	}
	defer container.Terminate(ctx)

	ip, err := container.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Error(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", ip, port.Port()),
	})
	defer rdb.Close()

	s := redis_store.New(rdb, 5*time.Minute)
	suite.Run(t, &store.TestSuite{Subject: s})
}
