---
sidebar_position: 9
title: MongoDB
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="MongoDB"
  description="MongoDB is a popular NoSQL document database that provides high performance, high availability, and easy scalability. It stores data in flexible, JSON-like documents with dynamic schemas."
  category="Document Database"
  packageName="backend/mongodb"
  importPath="github.com/kivigo/backends/mongodb"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "ACID Transactions", supported: true, description: "Multi-document transactions" },
    { name: "Replication", supported: true, description: "Master-slave and replica sets" },
    { name: "Sharding", supported: true, description: "Horizontal scaling across multiple servers" },
    { name: "Indexing", supported: true, description: "Compound, text, and geospatial indexes" },
    { name: "Aggregation Pipeline", supported: true, description: "Complex data processing and analytics" }
  ]}
  dependencies={[
    "MongoDB server (4.4 or later recommended)",
    "Network connectivity to MongoDB instance",
    "Valid connection string with authentication if required"
  ]}
  installationNotes="Requires a running MongoDB instance. You can use MongoDB Atlas (cloud), run locally, or use Docker: `docker run -d -p 27017:27017 mongo:latest`"
  configurationExample={`package main

import (
    "context"
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/mongodb"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Basic configuration
    opt := mongodb.NewOptions()
    opt.ConnectionURI = "mongodb://localhost:27017"
    opt.Database = "kivigo"
    opt.Collection = "keyvalue"

    // Advanced configuration with authentication
    authOpt := mongodb.NewOptions()
    authOpt.ConnectionURI = "mongodb://username:password@mongodb-server:27017/kivigo?authSource=admin"
    authOpt.Database = "kivigo"
    authOpt.Collection = "documents"
    
    // Optional: Configure connection pool and timeouts
    authOpt.MaxPoolSize = 100
    authOpt.ConnectTimeout = 10 * time.Second
    authOpt.SocketTimeout = 30 * time.Second
    
    // Optional: Configure write concern and read preference
    authOpt.WriteConcern = options.WriteConcern{
        W: "majority",
        J: true,
    }
    authOpt.ReadPreference = options.ReadPreference{
        Mode: "primaryPreferred",
    }
    
    // Create backend
    kvStore, err := mongodb.New(authOpt)
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
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/mongodb"
)

type DocumentMetadata struct {
    ID          string    \`json:"_id" bson:"_id"\`
    Title       string    \`json:"title" bson:"title"\`
    Author      string    \`json:"author" bson:"author"\`
    Tags        []string  \`json:"tags" bson:"tags"\`
    CreatedAt   time.Time \`json:"created_at" bson:"created_at"\`
    UpdatedAt   time.Time \`json:"updated_at" bson:"updated_at"\`
    Version     int       \`json:"version" bson:"version"\`
    Content     string    \`json:"content" bson:"content"\`
}

func main() {
    // Setup
    opt := mongodb.NewOptions()
    opt.ConnectionURI = "mongodb://localhost:27017"
    opt.Database = "kivigo"
    opt.Collection = "documents"

    kvStore, err := mongodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store document
    doc := DocumentMetadata{
        ID:        "doc-12345",
        Title:     "KiviGo Documentation",
        Author:    "Development Team",
        Tags:      []string{"go", "database", "kvstore", "documentation"},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Version:   1,
        Content:   "This is the comprehensive documentation for KiviGo...",
    }
    
    err = client.Set(ctx, "document:doc-12345", doc)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve document
    var retrievedDoc DocumentMetadata
    err = client.Get(ctx, "document:doc-12345", &retrievedDoc)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Document: %+v\\n", retrievedDoc)
    
    // Update document version
    retrievedDoc.Version++
    retrievedDoc.UpdatedAt = time.Now()
    retrievedDoc.Content = "Updated content for KiviGo documentation..."
    
    err = client.Set(ctx, "document:doc-12345", retrievedDoc)
    if err != nil {
        log.Fatal(err)
    }
    
    // List documents by prefix
    keys, err := client.List(ctx, "document:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Document Keys: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/kivigo/backends/mongodb"
)

func main() {
    opt := mongodb.NewOptions()
    opt.ConnectionURI = "mongodb://localhost:27017"
    opt.Database = "kivigo"
    opt.Collection = "keyvalue"

    kvStore, err := mongodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("MongoDB unhealthy: %v", err)
        return
    }
    
    log.Println("MongoDB is healthy")
    
    // Advanced health monitoring with replica set status
    ticker := time.NewTicker(45 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            start := time.Now()
            
            // Test write/read operation
            testKey := "health:check"
            testData := map[string]interface{}{
                "timestamp": time.Now().Unix(),
                "node":      "health-monitor",
                "check_id":  time.Now().Format("20060102-150405"),
            }
            
            if err := kvStore.SetRaw(ctx, testKey, []byte("health")); err != nil {
                log.Printf("Health check write failed: %v", err)
                continue
            }
            
            _, err := kvStore.GetRaw(ctx, testKey)
            if err != nil {
                log.Printf("Health check read failed: %v", err)
                continue
            }
            
            latency := time.Since(start)
            log.Printf("Health check passed (latency: %v)", latency)
            
            // Monitor connection pool
            if latency > 100*time.Millisecond {
                log.Printf("WARNING: High MongoDB latency detected: %v", latency)
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
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/mongodb"
)

func main() {
    opt := mongodb.NewOptions()
    opt.ConnectionURI = "mongodb://localhost:27017"
    opt.Database = "kivigo"
    opt.Collection = "analytics"

    kvStore, err := mongodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch insert analytics data
    analyticsData := map[string]interface{}{
        "analytics:page_views:2024-01": map[string]interface{}{
            "page": "/home", "views": 15420, "unique_visitors": 8932,
            "avg_time": "2m 34s", "bounce_rate": 0.32,
        },
        "analytics:page_views:2024-02": map[string]interface{}{
            "page": "/docs", "views": 9876, "unique_visitors": 5432,
            "avg_time": "4m 12s", "bounce_rate": 0.18,
        },
        "analytics:user_events:2024-01": map[string]interface{}{
            "event": "button_click", "count": 2341, "conversion_rate": 0.15,
            "locations": []string{"header", "footer", "sidebar"},
        },
        "analytics:performance:2024-01": map[string]interface{}{
            "avg_response_time": "245ms", "p95_response_time": "892ms",
            "error_rate": 0.02, "uptime": 0.9995,
        },
    }
    
    err = client.BatchSet(ctx, analyticsData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch retrieve analytics reports
    keys := []string{
        "analytics:page_views:2024-01",
        "analytics:page_views:2024-02",
        "analytics:user_events:2024-01",
        "analytics:performance:2024-01",
    }
    
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    for key, value := range results {
        fmt.Printf("%s: %v\\n", key, value)
    }
    
    // Batch archive old analytics data
    oldKeys := []string{"analytics:page_views:2023-12", "analytics:user_events:2023-12"}
    err = client.BatchDelete(ctx, oldKeys)
    if err != nil {
        log.Printf("Warning: Some old data deletion failed: %v", err)
    }
    
    fmt.Println("Analytics batch operations completed")
}`}
  notes={[
    "MongoDB stores documents in BSON format, which supports rich data types",
    "Use proper indexing strategies for frequently queried fields",
    "Consider sharding for horizontal scaling beyond single server capacity",
    "MongoDB transactions are available across replica sets and sharded clusters",
    "Use MongoDB Compass for visual database management and query optimization",
    "Monitor oplog size for replica sets to prevent synchronization issues",
    "GridFS can be used for storing files larger than 16MB document limit",
    "Connection pooling is essential for high-performance applications"
  ]}
  links={[
    { text: "MongoDB Official Documentation", url: "https://docs.mongodb.com/" },
    { text: "MongoDB Go Driver", url: "https://github.com/mongodb/mongo-go-driver" },
    { text: "MongoDB Atlas (Cloud)", url: "https://www.mongodb.com/cloud/atlas" },
    { text: "MongoDB Performance Best Practices", url: "https://docs.mongodb.com/manual/administration/analyzing-mongodb-performance/" }
  ]}
/>
