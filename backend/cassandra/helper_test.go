//go:build !unit

package cassandra

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
	hosts     []string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	ctx := context.Background()

	fmt.Println("🚀 Starting Cassandra (testcontainers)...")

	// Use an older, more stable Cassandra version for CI compatibility
	req := tc.ContainerRequest{
		Image:        "cassandra:3.11",
		ExposedPorts: []string{"9042/tcp"},
		WaitingFor:   wait.ForListeningPort("9042/tcp").WithStartupTimeout(60 * time.Second),
		Env: map[string]string{
			"CASSANDRA_CLUSTER_NAME": "Test Cluster",
			"CASSANDRA_DC":           "datacenter1",
			"CASSANDRA_RACK":         "rack1",
		},
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

	port, err := c.MappedPort(ctx, "9042")
	if err != nil {
		return nil, err
	}

	hosts := []string{fmt.Sprintf("%s:%s", host, port.Port())}

	r := &container{
		container: c,
		hosts:     hosts,
	}

	// Give Cassandra time to initialize
	time.Sleep(5 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Cassandra ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
