
<img align="left" width="250"  src="https://kivigo.github.io/img/logo-kivigo.png" alt="KiviGo Logo" />

[![Go Reference](https://pkg.go.dev/badge/github.com/kivigo/kivigo.svg)](https://pkg.go.dev/github.com/kivigo/kivigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/kivigo/kivigo)](https://goreportcard.com/report/github.com/kivigo/kivigo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)

KiviGo is a lightweight key-value store library for Go. It provides a simple interface for storing and retrieving key-value pairs, supporting multiple backends (E.g. Redis,  BoltDB) and encoders (JSON, YAML, etc.). KiviGo is designed to be easy to use, performant, and flexible.

**Why "KiviGo"?**  
The name is a play on words: "Kivi" sounds like "key-value" (the core concept of the library) and "Go" refers to the Go programming language. It also playfully evokes the fruit "kiwi" 🥝 !

## ✨ Features

- Unified interface for different backends ([Redis](pkg/backend/redis/redis.go), [local/BoltDB](pkg/backend/local/local.go), etc.)
- [Pluggable encoding/decoding (JSON, YAML, etc.)](https://github.com/kivigo/encoders)
- Health check support (with custom checks)
- List, add, and delete keys
- Easily extensible for new backends or encoders

## 📚 Documentation

Visit our comprehensive documentation at **[https://kivigo.github.io/kivigo/](https://kivigo.github.io/kivigo/)** for:

- **Getting Started Guide** - Set up your first key-value store in minutes
- **Complete Backend Documentation** - Detailed guides for all supported backends
- **Advanced Features** - Health checks, custom backends, encoders, batch operations, and more
- **Code Examples** - Real-world usage patterns and best practices

For API reference, see [pkg.go.dev/github.com/kivigo/kivigo](https://pkg.go.dev/github.com/kivigo/kivigo).

## 🥝 Motivation

KiviGo was created to simplify and unify key-value storage in Go applications. In many projects, developers face the challenge of switching between different storage backends (like Redis, BoltDB, or in-memory stores) or need to support multiple serialization formats (such as JSON or YAML). Each backend often comes with its own API, error handling, and setup, making code harder to maintain and test.

**What problem does KiviGo solve?**

- Provides a single, consistent API for key-value operations, regardless of the underlying backend.
- Makes it easy to swap storage engines (e.g., from local development with BoltDB to production with Redis) without changing your application logic.
- Supports pluggable encoders, so you can choose or implement the serialization format that fits your needs.
- Simplifies testing by providing a mock backend.
- Enables health checks and batch operations in a backend-agnostic way.

**Use cases:**

- Building microservices that need simple, fast key-value storage with the flexibility to change backends.
- Prototyping applications locally with BoltDB, then moving to Redis for production.
- Writing unit tests for storage logic without requiring a real database.
- Implementing custom backends (e.g., in-memory, cloud-based) while keeping the same client code.
- Supporting advanced features like batch operations, health checks, and custom serialization with minimal effort.

KiviGo helps you focus on your application logic, not on backend-specific details.

## 📊 Comparison: KiviGo vs Other Go Key-Value Libraries

> ⚠️ **Note:** The following comparison is provided for convenience and is based on the state of these libraries at the time of writing. Features and APIs may evolve over time—please refer to each project's documentation for the most up-to-date information.

There are several Go libraries for key-value storage, each with different goals and trade-offs. Here’s how KiviGo compares to some popular alternatives:

| Library         | Unified API | Pluggable Backends | Pluggable Encoders | Health Checks | Batch Ops | Mock/Test Support |
|-----------------|:----------:|:------------------:|:------------------:|:-------------:|:--------:|:-----------------:|
| **KiviGo**      | ✅         | ✅                 | ✅                 | ✅            | ✅       | ✅                |
| [gokv](https://github.com/philippgille/gokv)  | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| [libkv](https://github.com/docker/libkv) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| [gokvstores](https://github.com/ulule/gokvstores) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

KiviGo is designed for projects that need flexibility, testability, and the ability to swap storage or serialization strategies with minimal code changes.

### 🛠️ Backend Options Initialization

All KiviGo backends provide two helper functions for option management:

- **NewOptions()**  
  Returns an empty options struct for the backend.  
  Example:  

  ```go
  opts := backend.NewOptions() // All fields are zero values
  ```

- **DefaultOptions(...)**  
  Returns a recommended or minimal set of options for the backend.  
  This function can accept parameters to customize the defaults.  
  Example:  

  ```go
  opts := backend.DefaultOptions(path, otherParams...)
  ```

This design makes it easy to discover, configure, and override backend options in a consistent way across all supported backends.

---

**License**: MIT

---

## 🤝 Open Source & Contributions

KiviGo is an open source project. **Contributions, issues, and feature requests are welcome!**  
Feel free to open a pull request or an issue to help improve the project.
