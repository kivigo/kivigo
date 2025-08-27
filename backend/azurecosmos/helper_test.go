package azurecosmos

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
	endpoint  string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	ctx := context.Background()

	fmt.Println("🚀 Starting Azure Cosmos DB Emulator (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest",
		ExposedPorts: []string{"8081/tcp", "10251/tcp", "10252/tcp", "10253/tcp", "10254/tcp"},
		Env: map[string]string{
			"AZURE_COSMOS_EMULATOR_PARTITION_COUNT":         "10",
			"AZURE_COSMOS_EMULATOR_ENABLE_DATA_PERSISTENCE": "false",
		},
		WaitingFor: wait.ForListeningPort("8081/tcp").WithStartupTimeout(180 * time.Second),
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

	port, err := c.MappedPort(ctx, "8081")
	if err != nil {
		return nil, err
	}

	endpoint := "https://" + net.JoinHostPort(host, port.Port())

	r := &container{
		container: c,
		endpoint:  endpoint,
	}

	// Wait a bit for the emulator to fully initialize after the port is ready
	time.Sleep(5 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Azure Cosmos DB Emulator...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
