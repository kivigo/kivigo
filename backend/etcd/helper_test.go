package etcd

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
	endpoints []string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	fmt.Println("🚀 Starting etcd (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "quay.io/coreos/etcd:v3.5.14",
		ExposedPorts: []string{"2379/tcp"},
		WaitingFor:   wait.ForListeningPort("2379/tcp").WithStartupTimeout(30 * time.Second),
		Cmd:          []string{"etcd", "--advertise-client-urls", "http://0.0.0.0:2379", "--listen-client-urls", "http://0.0.0.0:2379"},
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

	port, err := c.MappedPort(context.Background(), "2379")
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port()) //nolint:nosprintfhostport

	r := &container{
		container: c,
		endpoints: []string{endpoint},
	}

	t.Setenv("ETCD_ENDPOINT", endpoint)
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping etcd ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
