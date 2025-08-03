/*
Package kivigo is a lightweight, extensible key-value store library for Go.

KiviGo provides a unified interface for storing and retrieving key-value pairs,
with support for multiple backends (such as Redis and BoltDB) and pluggable encoders (JSON, YAML, etc.).

Features:
  - Simple API for Get, Set, Delete, List operations
  - Multiple backends: Redis, BoltDB (local), and more
  - Pluggable encoding (JSON, YAML, ...)
  - Health checks (with custom checks)
  - Extensible: add your own backend or encoder
  - Batch operations and transactions (if supported by backend)
  - Key expiration (TTL) and hooks (events) [if supported]

See README.md for usage examples and backend-specific features.
*/
package kivigo
