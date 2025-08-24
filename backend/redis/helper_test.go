package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type container struct {
	container tc.Container
	addr      string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	fmt.Println("🚀 Starting Redis (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "redis:8",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}

	c, err := tc.GenericContainer(context.Background(), tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := c.Host(context.Background())
	if err != nil {
		return nil, err
	}

	port, err := c.MappedPort(context.Background(), "6379")
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	r := &container{
		container: c,
		addr:      addr,
	}
	// Set env for tests
	t.Setenv("REDIS_ADDR", addr)

	// Wait a bit for Redis to be ready
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Redis ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
