package azurecosmos

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
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
		WaitingFor: wait.ForListeningPort("8081/tcp").WithStartupTimeout(120 * time.Second),
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

	// Wait for Cosmos DB Emulator to be ready by polling the endpoint
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Required for emulator
		},
	}

	ready := false
	maxWait := 30 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			ready = true
			resp.Body.Close()

			break
		}

		time.Sleep(1 * time.Second)
	}

	if !ready {
		return nil, fmt.Errorf("Cosmos DB Emulator not ready after %v", maxWait)
	}

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Azure Cosmos DB Emulator...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
