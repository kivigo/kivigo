# KiviGo - Key-Value Store Library for Go

**ALWAYS follow these instructions first and fallback to additional search and context gathering only if the information here is incomplete or found to be in error.**

KiviGo is a lightweight, modular key-value store library for Go that provides a unified interface for different backends (Redis, BoltDB, Consul, etcd, Badger) and encoders (JSON, YAML). Each backend is implemented as a separate Go module to minimize dependencies.

## Working Effectively

### Bootstrap and Build
Always run these commands in sequence to set up the development environment:

```bash
# 1. Download dependencies (takes ~1-2 seconds)
go mod download

# 2. Build main package (takes ~7 seconds)
go build ./pkg/...

# 3. Install golangci-lint for linting
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Testing Strategy
The project uses a modular testing approach:

```bash
# Test main package (takes ~3 seconds)
go test ./pkg/...

# Test all backends individually (takes 2-10 seconds each) - NEVER CANCEL
for backend in backend/*/; do
    if [[ -f "$backend/go.mod" ]]; then
        echo "Testing $(basename "$backend")..."
        cd "$backend" && go test ./... && cd ..
    fi
done
```

**CRITICAL TIMING**: Backend tests may take up to 6 seconds each due to external dependencies (Redis, etcd, Consul require Docker containers). First run includes module downloads. Set timeout to 300+ seconds. NEVER CANCEL during backend testing.

### Linting
Always run linting before committing (takes ~3 seconds):
```bash
golangci-lint run --timeout=5m
```

Note: You may see deprecation warnings for `wsl` linter - these are harmless and do not affect the build.

## Validation Scenarios

### ALWAYS run these validation steps after making changes:

1. **Basic Functionality Test** - Create and run this validation:
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/pkg/encoder"
    "github.com/azrod/kivigo/pkg/mock"
)

func main() {
    // Test with mock backend
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
    if err != nil {
        log.Fatal(err)
    }

    // Store and retrieve a value
    if err := c.Set(context.Background(), "test-key", "test-value"); err != nil {
        log.Fatal(err)
    }

    var value string
    if err := c.Get(context.Background(), "test-key", &value); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✅ KiviGo functional test passed: stored and retrieved '%s'\n", value)
}
```

2. **Complete Validation Script** - Run this to validate all functionality:
```bash
#!/bin/bash
set -e

echo "🧪 KiviGo Validation Tests"
cd /home/runner/work/kivigo/kivigo

# Build and test main package
time go build ./pkg/...
time go test ./pkg/...

# Test all backends
for backend in backend/*/; do
    if [[ -f "$backend/go.mod" ]]; then
        echo "Testing $(basename "$backend")..."
        (cd "$backend" && go test ./...)
    fi
done

# Run functional validation
go run /path/to/validation/script.go

# Lint
export PATH=$PATH:$(go env GOPATH)/bin
golangci-lint run --timeout=5m

echo "✅ All validation tests passed!"
```

## Architecture Understanding

### Module Structure
- **Main package**: `pkg/` contains core client, encoders, and models
- **Backends**: Each `backend/*/` directory is a separate Go module with its own dependencies
- **Examples**: `examples/` contains usage patterns (may have dependency issues in development)

### Key Directories
```
pkg/client/     # Main client implementation
pkg/encoder/    # JSON, YAML encoders
pkg/mock/       # Mock backend for testing
backend/badger/ # BadgerDB backend
backend/redis/  # Redis backend  
backend/etcd/   # etcd backend
backend/consul/ # Consul backend
backend/local/  # BoltDB backend
```

### Common Commands Reference
```bash
# Repository root contents
ls -la
# .github/ .golangci.yml CONTRIBUTING.md README.md backend/ examples/ pkg/ go.mod

# Available backends
ls backend/
# badger consul etcd local redis

# Core packages
ls pkg/
# client encoder errs mock models
```

## Timing Expectations and Timeouts

**NEVER CANCEL** any of these operations. Always set appropriate timeouts:

- `go mod download`: ~1 second when cached, 5-10 seconds fresh (timeout: 120s)
- `go build ./pkg/...`: ~5 seconds (timeout: 300s)  
- `go test ./pkg/...`: ~2.5 seconds (timeout: 300s)
- Backend tests: 3-6 seconds each with downloads, <1 second cached (timeout: 300s)
- `golangci-lint run`: ~1 second (timeout: 300s)
- Full validation suite: ~1.5 minutes total (timeout: 600s)

## Requirements

- **Go version**: 1.23+ (project uses Go 1.23.8 with toolchain 1.24.5)
- **golangci-lint**: Install using: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest`
- **Docker**: Some backend tests (etcd, consul) use testcontainers - may fail in environments without Docker
- **Environment**: Ensure `export PATH=$PATH:$(go env GOPATH)/bin` for golangci-lint access

## Troubleshooting

### Build Issues
- If build fails, ensure Go 1.23+ is installed
- Run `go mod tidy` in the root directory if dependencies are missing

### Backend Test Issues  
- etcd/consul tests need Docker - may fail in environments without Docker
- badger/redis/local backends work without external dependencies
- Use mock backend for reliable testing scenarios

### Example Issues
- **Examples in `examples/` have module dependency issues** - these reference published versions, not local development code
- If you try `cd examples/local && go run main.go`, you'll get: `undefined: models.Backend`
- **Solution**: Use the provided validation scripts instead of running examples directly
- For testing functionality, always use the mock backend validation script provided

### Linting Issues
- Deprecation warnings for `wsl` linter are harmless
- Update `.golangci.yml` to use `wsl_v5` if needed
- All code should pass linting with 0 issues

## Development Workflow

1. **Make changes** to code in `pkg/` or `backend/*/`
2. **Build and test** using the bootstrap commands
3. **Validate** with functional test script
4. **Lint** with golangci-lint
5. **Commit** only after all validation passes

ALWAYS test both the main package and relevant backend modules when making changes. The modular architecture means changes to core interfaces may affect multiple backends.

## Quick Reference Commands

```bash
# Complete fresh setup and validation (copy-paste ready)
cd /path/to/kivigo
go mod download
go build ./pkg/...
go test ./pkg/...
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
export PATH=$PATH:$(go env GOPATH)/bin
golangci-lint run --timeout=5m

# Test all backends
for backend in backend/*/; do
    if [[ -f "$backend/go.mod" ]]; then
        echo "Testing $(basename "$backend")..."
        (cd "$backend" && go test ./...)
    fi
done

# Quick functional validation
cat > test_kivigo.go << 'EOF'
package main
import (
    "context"; "fmt"
    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/pkg/encoder"
    "github.com/azrod/kivigo/pkg/mock"
)
func main() {
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    c, _ := client.New(mockKV, client.Option{Encoder: encoder.JSON})
    c.Set(context.Background(), "test", "value")
    var v string; c.Get(context.Background(), "test", &v)
    fmt.Printf("✅ Test passed: %s\n", v)
}
EOF
go run test_kivigo.go && rm test_kivigo.go
```