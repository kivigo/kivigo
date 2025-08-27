# Selective CI Testing and Linting

This document describes the optimized CI workflows that run tests and linting only on modules that have actually changed.

## Overview

KiviGo now uses intelligent change detection to optimize CI performance by running tests and linting only on the modules that have been modified in a pull request.

## Workflows

### go-test.yml - Selective Testing

**Triggers**: Pull requests (opened, synchronize, reopened)

**Change Detection Logic**:
- **Main package**: Runs tests when files in `pkg/`, root `*.go` files, or `go.mod`/`go.sum` are modified
- **Backend modules**: Runs tests only on backends where files in `backend/{name}/` are modified

**Benefits**:
- ⚡ Faster CI: Only tests changed modules instead of everything
- 📊 Maintained coverage: Full Coveralls integration preserved  
- 🎯 Focused feedback: Clear reporting on which modules were tested

### go-lint.yml - Selective Linting

**Triggers**: Pull requests (opened, synchronize, reopened)

**Change Detection Logic**:
- **Main package**: Runs linting when files in `pkg/`, root `*.go` files, or `go.mod`/`go.sum` are modified
- **Backend modules**: Runs linting only on backends where files in `backend/{name}/` are modified

**Benefits**:
- ⚡ Faster CI: Only lint changed modules
- 🎯 Focused feedback: Clear reporting on which modules were linted
- ✅ Same quality: Maintains all existing linting rules

## Coverage Reporting

### Coveralls Integration

- **Parallel reporting**: Each module uploads coverage independently
- **Aggregation**: Coverage is properly aggregated across all tested modules
- **PR links**: Test summary includes direct links to Coveralls reports

### PR Summary Reports

Both workflows generate comprehensive summary reports that show:
- Which modules had changes detected
- Which tests/linting jobs ran or were skipped
- Links to coverage reports (when tests ran)
- Clear status indicators (✅ passed, ❌ failed, ⏭️ skipped)

## Example Scenarios

### Scenario 1: Core Package Changes Only
```
Files changed: pkg/client/client.go, pkg/encoder/json.go
Result: Only main package tests and linting run
Time saved: ~3-5 minutes (skipping 6+ backend tests)
```

### Scenario 2: Single Backend Changes
```
Files changed: backend/redis/redis.go, backend/redis/redis_test.go  
Result: Only Redis backend tests and linting run
Time saved: ~5-7 minutes (skipping main + 5 other backends)
```

### Scenario 3: Documentation Changes
```
Files changed: README.md, docs/api.md
Result: All tests and linting skipped
Time saved: ~8-10 minutes (entire test suite)
```

### Scenario 4: Mixed Changes
```
Files changed: pkg/client.go, backend/redis/redis.go, go.mod
Result: Main package + Redis backend tests and linting run  
Time saved: ~3-5 minutes (skipping 5 other backends)
```

## Integration with Existing Workflows

The selective CI workflows integrate seamlessly with existing KiviGo processes:

- **Dependabot PRs**: Continue using the specialized `dependency-updates.yml` workflow
- **Regular PRs**: Use the new selective `go-test.yml` and `go-lint.yml` workflows
- **Coverage**: Maintains full Coveralls integration and historical data
- **Quality**: Same test and lint standards applied to changed modules

## Backward Compatibility

- ✅ No breaking changes to existing coverage reporting
- ✅ Same test and lint quality standards maintained
- ✅ All existing integrations (Coveralls, etc.) preserved
- ✅ Works with existing development workflows

## Performance Impact

Expected performance improvements:

| Scenario | Before | After | Time Saved |
|----------|--------|-------|------------|
| Core changes only | ~8-10 min | ~3-4 min | 50-60% |
| Single backend | ~8-10 min | ~2-3 min | 70-75% |
| Documentation only | ~8-10 min | ~30 sec | 95% |
| Mixed changes | ~8-10 min | ~4-6 min | 40-50% |

*Times are approximate and depend on backend complexity and test coverage.*