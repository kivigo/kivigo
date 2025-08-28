package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type container struct {
	container tc.Container
	dsn       string
}

func start(t *testing.T) (*container, error) {
	t.Helper()

	ctx := context.Background()

	fmt.Println("🚀 Starting PostgreSQL (testcontainers)...")

	// PostgreSQL container configuration
	req := tc.ContainerRequest{
		Image:        "postgres:alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "kivigo",
			"POSTGRES_USER":     "testuser",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(60*time.Second),
		),
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

	port, err := c.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s/kivigo?sslmode=disable", net.JoinHostPort(host, port.Port()))

	r := &container{
		container: c,
		dsn:       dsn,
	}

	// Wait a bit for PostgreSQL to be fully ready
	time.Sleep(3 * time.Second)

	// Verify connection works
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open test connection: %w", err)
	}
	defer db.Close()

	// Try to ping the database with retries
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		if i == 9 {
			return nil, fmt.Errorf("failed to connect to PostgreSQL after retries")
		}
		time.Sleep(1 * time.Second)
	}

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping PostgreSQL ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
