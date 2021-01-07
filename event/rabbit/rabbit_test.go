package rabbit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/event/rabbit"
)

func TestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping event/rabbit test")
	}

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3.8.9-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForListeningPort("5672/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	ip, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5672")
	require.NoError(t, err)

	subject, err := rabbit.New(fmt.Sprintf("amqp://guest:guest@%s:%s/", ip, port.Port()))
	require.NoError(t, err)

	suite.Run(t, &event.TestSuite{
		S: subject,
		E: subject,
	})
}
