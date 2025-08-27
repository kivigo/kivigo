package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
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

	fmt.Println("🚀 Starting MySQL (testcontainers)...")

	// MySQL container configuration
	req := tc.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "rootpassword",
			"MYSQL_DATABASE":      "kivigo",
			"MYSQL_USER":          "testuser",
			"MYSQL_PASSWORD":      "testpass",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("3306/tcp"),
			wait.ForLog("ready for connections").WithStartupTimeout(60*time.Second),
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

	port, err := c.MappedPort(ctx, "3306")
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("testuser:testpass@tcp(%s:%s)/kivigo", host, port.Port())

	r := &container{
		container: c,
		dsn:       dsn,
	}

	// Wait a bit for MySQL to be fully ready
	time.Sleep(3 * time.Second)

	// Verify connection works
	db, err := sql.Open("mysql", dsn)
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
			return nil, fmt.Errorf("failed to connect to MySQL after retries")
		}
		time.Sleep(1 * time.Second)
	}

	return r, nil
}

func (c *container) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping MySQL ...")

	if c.container != nil {
		return c.container.Terminate(ctx)
	}

	return nil
}
