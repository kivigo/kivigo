package consul

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

	ctx := context.Background()

	fmt.Println("🚀 Starting Consul (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "hashicorp/consul:latest",
		ExposedPorts: []string{"8500/tcp"},
		WaitingFor:   wait.ForListeningPort("8500/tcp").WithStartupTimeout(30 * time.Second),
	}

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := c.MappedPort(ctx, "8500")
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	r := &container{
		container: c,
		addr:      addr,
	}
	// Set env for tests
	t.Setenv("CONSUL_ADDR", addr)

	// Wait a bit for Consul to be ready
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Consul ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
