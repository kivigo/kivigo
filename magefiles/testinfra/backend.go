package testinfra

import "context"

// BackendInfra standardizes the lifecycle and registration of test backends.
type BackendInfra interface {
	// Name returns the backend name (e.g. "redis", "etcd").
	Name() string
	// Start launches the backend service (e.g. docker run ...).
	Start(ctx context.Context) error
	// Stop stops the backend service (e.g. docker stop ...).
	Stop(ctx context.Context) error
	// Register registers the backend in a global registry for test orchestration.
	Register()

	// Tests is a list of test cases that this backend supports.
	Tests() []string
}
