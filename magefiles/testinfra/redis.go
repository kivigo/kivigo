package testinfra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

var _ BackendInfra = (*BRedis)(nil)

func init() { //nolint:gochecknoinits
	// Register the Redis backend on package initialization
	var redis BRedis

	redis.Register()
}

// BRedis implements BackendInfra for Redis.
type BRedis struct{}

func (r *BRedis) Tests() []string {
	return []string{
		"TestRedis_BasicOps",
	}
}

func (r *BRedis) Name() string {
	return "redis"
}

func (r *BRedis) Start(ctx context.Context) error {
	fmt.Println("🚀 Starting Redis (docker)...")

	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "--rm", "-p", "6379:6379", "--name", "kivigo-redis", "redis:8")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r *BRedis) Stop(ctx context.Context) error {
	fmt.Println("🛑 Stopping Redis (docker)...")

	cmd := exec.CommandContext(ctx, "docker", "stop", "kivigo-redis")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r *BRedis) Register() {
	RegisterBackend(r)
}
