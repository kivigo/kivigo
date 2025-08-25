package memcached

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

	fmt.Println("🚀 Starting Memcached (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "memcached:1.6-alpine",
		ExposedPorts: []string{"11211/tcp"},
		WaitingFor:   wait.ForListeningPort("11211/tcp").WithStartupTimeout(30 * time.Second),
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

	port, err := c.MappedPort(context.Background(), "11211")
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	r := &container{
		container: c,
		addr:      addr,
	}
	// Set env for tests
	t.Setenv("MEMCACHED_ADDR", addr)

	// Wait a bit for Memcached to be ready
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Memcached ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
