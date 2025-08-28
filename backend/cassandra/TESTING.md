# Cassandra Backend Testing

This document explains the testing setup for the Cassandra backend and how to handle CI environment issues.

## Test Categories

### Unit Tests
- **Run with**: `go test -tags=unit`
- **Requirements**: None (no external dependencies)
- **Purpose**: Tests configuration, options, and basic error handling

### Integration Tests  
- **Run with**: `go test` (default)
- **Requirements**: Docker (uses testcontainers)
- **Purpose**: Tests full Cassandra integration with real database

## Environment Variables

### `SKIP_TESTCONTAINERS=true`
Skips all testcontainer-based tests. Useful in CI environments where Docker may not be available or may have issues.

```bash
# Skip integration tests, run only unit tests
SKIP_TESTCONTAINERS=true go test ./...

# Or run unit tests explicitly  
go test -tags=unit ./...
```

### CI Environment Detection
The tests automatically detect CI environments via the `CI` environment variable and adjust behavior accordingly.

## Testcontainer Configuration

The Cassandra backend uses:
- **Image**: `cassandra:3.11` (stable version for CI compatibility)
- **Port**: `9042/tcp` 
- **Startup Timeout**: 60 seconds
- **Additional Wait**: 5 seconds for schema initialization

## Common CI Issues

### Container Startup Timeout
If tests fail with Ryuk or container startup timeouts:

```bash
# Option 1: Skip testcontainer tests
SKIP_TESTCONTAINERS=true go test ./...

# Option 2: Run only unit tests
go test -tags=unit ./...
```

### Docker Permission Issues
Ensure the CI runner has Docker access, or use the skip options above.

### Memory/Resource Constraints
Cassandra requires significant memory. If CI environments have resource constraints, use unit tests only.

## Build Tags

- `unit`: Compiles only unit tests (no testcontainers)
- `!unit`: Default - includes integration tests

## Example CI Configuration

```yaml
# GitHub Actions example
- name: Test Cassandra Backend
  run: |
    cd backend/cassandra
    # Try integration tests first
    if ! go test -timeout=300s ./...; then
      echo "Integration tests failed, running unit tests only"
      SKIP_TESTCONTAINERS=true go test -tags=unit ./...
    fi
```