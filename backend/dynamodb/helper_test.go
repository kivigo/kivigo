package dynamodb

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

	fmt.Println("🚀 Starting DynamoDB Local (testcontainers)...")

	req := tc.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		ExposedPorts: []string{"8000/tcp"},
		Cmd:          []string{"-jar", "DynamoDBLocal.jar", "-inMemory", "-sharedDb"},
		WaitingFor:   wait.ForListeningPort("8000/tcp").WithStartupTimeout(30 * time.Second),
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

	port, err := c.MappedPort(ctx, "8000")
	if err != nil {
		return nil, err
	}

	endpoint := "http://" + net.JoinHostPort(host, port.Port())

	r := &container{
		container: c,
		endpoint:  endpoint,
	}

	// Set env for tests
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("DYNAMODB_ENDPOINT", endpoint)

	// Wait a bit for DynamoDB Local to be ready
	time.Sleep(2 * time.Second)

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping DynamoDB Local ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
