---
sidebar_position: 2
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="BadgerDB"
  description="BadgerDB is a fast, embedded key-value database written in Go. It's designed for high-performance read and write operations with built-in caching and efficient memory usage."
  category="Embedded Store"
  packageName="backend/badger"
  importPath="github.com/kivigo/backends/badger"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "Persistence", supported: true, description: "Data is stored on disk" },
    { name: "ACID Transactions", supported: true, description: "Full ACID compliance" },
    { name: "Compression", supported: true, description: "Built-in data compression" },
    { name: "Concurrent Access", supported: true, description: "Thread-safe operations" }
  ]}
  dependencies={[
    "No external services required",
    "Disk space for data storage",
    "File system write permissions"
  ]}
  installationNotes="BadgerDB is an embedded database that stores data locally in files. No external services are required."
  configurationExample={`package main

import (
    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/badger"
)

func main() {
    // Basic configuration
    opt := badger.DefaultOptions("./data")

    // Custom configuration
    customOpt := badger.NewOptions()
    customOpt.Dir = "./data/keys"
    customOpt.ValueDir = "./data/values"
    customOpt.SyncWrites = true
    customOpt.MemTableSize = 64 << 20 // 64MB
    
    // Create backend
    kvStore, err := badger.New(customOpt)
    if err != nil {
        panic(err)
    }
    defer kvStore.Close()
    
    // Create client
    client, err := kivigo.New(kvStore)
    if err != nil {
        panic(err)
    }
}`}
  usageExample={`package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/badger"
)

type Config struct {
    AppName string \`json:"app_name"\`
    Port    int    \`json:"port"\`
    Debug   bool   \`json:"debug"\`
}

func main() {
    // Setup
    opt := badger.DefaultOptions("./data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store configuration
    config := Config{
        AppName: "MyApp",
        Port:    8080,
        Debug:   true,
    }
    
    err = client.Set(ctx, "app:config", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve configuration
    var retrievedConfig Config
    err = client.Get(ctx, "app:config", &retrievedConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Config: %+v\\n", retrievedConfig)
    
    // List keys with prefix
    keys, err := client.List(ctx, "app:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Keys: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/kivigo/backends/badger"
)

func main() {
    opt := badger.DefaultOptions("./data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("BadgerDB unhealthy: %v", err)
        return
    }
    
    log.Println("BadgerDB is healthy")
    
    // Periodic health monitoring
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := kvStore.Health(ctx); err != nil {
                log.Printf("Health check failed: %v", err)
            } else {
                log.Println("Health check passed")
            }
        case <-ctx.Done():
            return
        }
    }
}`}
  batchExample={`package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/badger"
)

func main() {
    opt := badger.DefaultOptions("./data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch set multiple values
    batchData := map[string]interface{}{
        "user:1": map[string]string{"name": "Alice", "email": "alice@example.com"},
        "user:2": map[string]string{"name": "Bob", "email": "bob@example.com"},
        "user:3": map[string]string{"name": "Charlie", "email": "charlie@example.com"},
    }
    
    err = client.BatchSet(ctx, batchData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch get multiple values
    keys := []string{"user:1", "user:2", "user:3"}
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    for key, value := range results {
        fmt.Printf("%s: %v\\n", key, value)
    }
    
    // Batch delete
    err = client.BatchDelete(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Batch operations completed")
}`}
  notes={[
    "BadgerDB creates two directories: one for keys and one for values",
    "The database automatically handles compaction and garbage collection",
    "For production use, consider tuning MemTableSize and other options based on your workload",
    "BadgerDB uses memory-mapped files for efficient access",
    "The database supports encryption at rest (configure via BadgerDB options)",
    "Transaction size limits apply - very large transactions may need to be split",
    "BadgerDB is optimized for SSD storage but works on traditional HDDs"
  ]}
  links={[
    { text: "BadgerDB Official Documentation", url: "https://dgraph.io/docs/badger/" },
    { text: "BadgerDB GitHub Repository", url: "https://github.com/dgraph-io/badger" },
    { text: "Performance Tuning Guide", url: "https://dgraph.io/docs/badger/get-started/#tuning-badgerdb" }
  ]}
/>
