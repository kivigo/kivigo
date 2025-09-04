## KiviGo v1.5.1 — Release Notes

### 🚀 What's Changed Since v1.5.0

#### 🏗️ Internal/Engineering
- **New scripts for contributors & CI:**
  - Added `scripts/lint.sh`: One-command linting with `golangci-lint --fix` for the repo and all backend modules.
  - Added `scripts/unit-test.sh`: Run unit tests with coverage for core and each backend, merges coverage profiles and generates HTML reports.
  - Added `scripts/merge_coverage.go`: Utility for merging multiple Go coverage profiles deterministically.

#### 🧪 Testing & Mock Improvements
- **Mock backend & client:**
  - Improved mock error handling for empty keys, not found, and batch operations.
  - Mock now implements stricter error returns for batch operations (empty batch, not found).
  - Tests for client and mock now cover error cases for key existence, batch edge cases, and deletion.
  - Added new methods and tests for key existence checks (`HasKey`, `HasKeys`, `MatchKeys`).

#### 🧹 API & Error Handling
- **Unified error codes:**
  - Error types from `pkg/errs/backend.go` and `pkg/errs/kv.go` are merged into a single source (`pkg/errs/errs.go`), removing redundant definitions.
  - All backends now use unified error types, e.g. `ErrNotFound`, `ErrEmptyKey`, `ErrEmptyBatch`, `ErrHealthCheckFailed`, etc.
  - Error codes in client and mock updated for consistency.
  - More robust error propagation for empty batch, key not found, and empty key scenarios.

#### 🛠️ Backend Consistency
- **Backends updated:**
  - Consul, etcd, DynamoDB, Local, Memcached, MongoDB, MySQL, PostgreSQL, Redis, Badger: Unified error return for health and batch methods.
  - Health check errors now propagate the actual backend error rather than wrapping.
  - Batch methods now reject empty batches with a consistent error.

#### 📚 Documentation & DX
- **Examples and Docs:**
  - New and improved usage examples in documentation (Getting Started, Examples, Operations).
  - Sidebar and versioned docs improved for easier navigation.
  - Docusaurus `lastVersion` updated to 1.5.0.
  - Algolia search and robots.txt added for docs search.

#### 🔧 Other
- **TODO.md and redundant files removed.**
- **Improved test and lint coverage.**

---

### How to Upgrade

```bash
# Core library
go get github.com/azrod/kivigo@v1.5.1

# Backends (example)
go get github.com/azrod/kivigo/backend/redis@v1.5.1
go get github.com/azrod/kivigo/backend/postgresql@v1.5.1
# ...and so on for each backend
```

---

### Full Diff

See all changes: [v1.5.0...v1.5.1](https://github.com/azrod/kivigo/compare/v1.5.0...v1.5.1)

---