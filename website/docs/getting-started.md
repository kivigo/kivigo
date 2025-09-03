---
sidebar_position: 2
---

# Getting Started

Welcome to KiviGo! This guide will help you get up and running with KiviGo's key-value store library in just a few minutes.

## What is KiviGo?

KiviGo is a lightweight, modular key-value store library for Go that provides a unified interface for different backends (Redis, BadgerDB, Consul, etcd, etc.) and encoders (JSON, YAML). Each backend is implemented as a separate Go module to minimize dependencies.

## Why KiviGo?

- **🔄 Unified Interface**: Switch between backends without changing your code
- **📦 Modular Design**: Each backend is a separate module - only install what you need
- **🎯 Type Safety**: Automatic marshaling/unmarshaling of Go types
- **🧪 Testing Ready**: Built-in mock backend for unit testing
- **⚡ Performance**: Optimized for both embedded and distributed scenarios
- **🛡️ Reliable**: Comprehensive error handling and health checks

## Quick Overview

Here's what working with KiviGo looks like:

```go
// Setup is backend-specific
client, err := kivigo.New(backend)

// But usage is always the same
err = client.Set(ctx, "key", value)    // Store any Go type
err = client.Get(ctx, "key", &value)   // Retrieve into any Go type
keys, err := client.List(ctx, prefix)  // List keys by prefix
err = client.Delete(ctx, "key")        // Remove data
```

## Getting Started Guide

Follow these steps to start using KiviGo:

### 1. 📦 [Installation](./installation)
Learn how to install KiviGo core and choose the right backend for your needs.

### 2. 🚀 [Quick Start](./quick-start) 
Get up and running with basic examples and understand the fundamentals.

### 3. ⚙️ [Operations](./operations)
Master all available operations: Set, Get, List, Delete, and more.

### 4. 💡 [Examples](./examples)
See practical, real-world examples including configuration management, caching, and session storage.

## Choose Your Path

Depending on your experience and needs:

**🔰 New to KiviGo?**
Start with [Installation](./installation) → [Quick Start](./quick-start)

**🎯 Want specific operations?**
Jump to [Operations](./operations) for comprehensive operation guides

**💼 Looking for real examples?**
Check out [Examples](./examples) for practical use cases

**🗄️ Need a specific backend?**
Browse [Backends](./backends/overview) to find the perfect storage solution

**🔧 Want advanced features?**
Explore [Advanced Features](./advanced/health-checks) for custom backends, batch operations, and more

## Community & Support

- **📚 Documentation**: Complete guides and API reference at [kivigo.dev](https://azrod.github.io/kivigo/)
- **🐛 Issues**: Report bugs and request features on [GitHub](https://github.com/azrod/kivigo/issues)
- **💬 Discussions**: Ask questions and share ideas in [GitHub Discussions](https://github.com/azrod/kivigo/discussions)

## What's Next?

Ready to start? Begin with **[Installation](./installation)** to set up KiviGo in your project!
