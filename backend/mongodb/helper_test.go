package mongodb

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type container struct {
	container tc.Container
	uri       string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	ctx := context.Background()

	fmt.Println("🚀 Starting MongoDB (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(30 * time.Second),
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

	port, err := c.MappedPort(ctx, "27017")
	if err != nil {
		return nil, err
	}

	uri := "mongodb://" + net.JoinHostPort(host, port.Port())

	r := &container{
		container: c,
		uri:       uri,
	}

	// Wait a bit for MongoDB to be ready
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping MongoDB ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
