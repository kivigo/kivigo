---
sidebar_position: 12
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="Azure Cosmos DB"
  description="Azure Cosmos DB is Microsoft's globally distributed, multi-model database service. It provides low-latency, elastic scale, and comprehensive SLAs for throughput, latency, availability, and consistency."
  category="Cloud Database"
  packageName="backend/azurecosmos"
  importPath="github.com/azrod/kivigo/backend/azurecosmos"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "Global Distribution", supported: true, description: "Multi-region replication" },
    { name: "Multiple Consistency Levels", supported: true, description: "From strong to eventual consistency" },
    { name: "Automatic Scaling", supported: true, description: "Elastic throughput scaling" },
    { name: "SLA Guarantees", supported: true, description: "99.999% availability SLA" },
    { name: "Multi-Model Support", supported: true, description: "Document, key-value, graph, and column-family" }
  ]}
  dependencies={[
    "Azure Cosmos DB account",
    "Valid connection string or account key",
    "Network connectivity to Azure"
  ]}
  installationNotes="Requires an active Azure Cosmos DB account. You can create a free tier account on Azure Portal."
  configurationExample={`package main

import (
    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/backend/azurecosmos"
)

func main() {
    // Configuration with connection string
    opt := azurecosmos.NewOptions()
    opt.ConnectionString = "AccountEndpoint=https://your-account.documents.azure.com:443/;AccountKey=your-key;"
    opt.DatabaseName = "kivigo"
    opt.ContainerName = "keyvalue"
    
    // Optional: Configure consistency level
    opt.ConsistencyLevel = azurecosmos.SessionConsistency
    
    // Optional: Configure request options
    opt.RequestOptions = azurecosmos.RequestOptions{
        RequestUnitsPerSecond: 400,
        PartitionKey:         "/id",
    }
    
    // Create backend
    kvStore, err := azurecosmos.New(opt)
    if err != nil {
        panic(err)
    }
    defer kvStore.Close()
    
    // Create client
    client, err := client.New(kvStore, client.Option{})
    if err != nil {
        panic(err)
    }
}`}
  usageExample={`package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/backend/azurecosmos"
)

type UserProfile struct {
    ID       string \`json:"id"\`
    Name     string \`json:"name"\`
    Email    string \`json:"email"\`
    Location string \`json:"location"\`
}

func main() {
    // Setup
    opt := azurecosmos.NewOptions()
    opt.ConnectionString = "AccountEndpoint=https://your-account.documents.azure.com:443/;AccountKey=your-key;"
    opt.DatabaseName = "kivigo"
    opt.ContainerName = "profiles"
    
    kvStore, err := azurecosmos.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := client.New(kvStore, client.Option{})
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store user profile
    profile := UserProfile{
        ID:       "user-123",
        Name:     "Alice Johnson",
        Email:    "alice@example.com",
        Location: "Seattle, WA",
    }
    
    err = client.Set(ctx, "profile:user-123", profile)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve user profile
    var retrievedProfile UserProfile
    err = client.Get(ctx, "profile:user-123", &retrievedProfile)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Profile: %+v\\n", retrievedProfile)
    
    // List keys with prefix
    keys, err := client.List(ctx, "profile:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Profile Keys: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"
    
    "github.com/azrod/kivigo/backend/azurecosmos"
)

func main() {
    opt := azurecosmos.NewOptions()
    opt.ConnectionString = "AccountEndpoint=https://your-account.documents.azure.com:443/;AccountKey=your-key;"
    opt.DatabaseName = "kivigo"
    opt.ContainerName = "keyvalue"
    
    kvStore, err := azurecosmos.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("Azure Cosmos DB unhealthy: %v", err)
        return
    }
    
    log.Println("Azure Cosmos DB is healthy")
    
    // Periodic health monitoring with circuit breaker pattern
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    consecutiveFailures := 0
    maxFailures := 3
    
    for {
        select {
        case <-ticker.C:
            if err := kvStore.Health(ctx); err != nil {
                consecutiveFailures++
                log.Printf("Health check failed (%d/%d): %v", consecutiveFailures, maxFailures, err)
                
                if consecutiveFailures >= maxFailures {
                    log.Println("Circuit breaker opened - service unavailable")
                    return
                }
            } else {
                consecutiveFailures = 0
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
    
    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/backend/azurecosmos"
)

func main() {
    opt := azurecosmos.NewOptions()
    opt.ConnectionString = "AccountEndpoint=https://your-account.documents.azure.com:443/;AccountKey=your-key;"
    opt.DatabaseName = "kivigo"
    opt.ContainerName = "keyvalue"
    
    kvStore, err := azurecosmos.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := client.New(kvStore, client.Option{})
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch set multiple documents
    batchData := map[string]interface{}{
        "product:1": map[string]interface{}{"name": "Laptop", "price": 999.99, "category": "electronics"},
        "product:2": map[string]interface{}{"name": "Phone", "price": 599.99, "category": "electronics"},
        "product:3": map[string]interface{}{"name": "Book", "price": 19.99, "category": "education"},
    }
    
    err = client.BatchSet(ctx, batchData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch get multiple documents
    keys := []string{"product:1", "product:2", "product:3"}
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
    
    fmt.Println("Batch operations completed successfully")
}`}
  notes={[
    "Azure Cosmos DB charges based on Request Units (RUs) consumed",
    "Consider using session consistency for most applications to balance performance and consistency",
    "Partition key design is crucial for performance - choose a key with good distribution",
    "Use batch operations to reduce RU consumption for multiple operations",
    "Monitor your RU consumption in Azure Portal to optimize costs",
    "Consider using autoscale throughput for unpredictable workloads",
    "The free tier provides 1000 RU/s and 25GB storage at no cost"
  ]}
  links={[
    { text: "Azure Cosmos DB Documentation", url: "https://docs.microsoft.com/en-us/azure/cosmos-db/" },
    { text: "Cosmos DB Go SDK", url: "https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/data/azcosmos" },
    { text: "Partitioning and Performance", url: "https://docs.microsoft.com/en-us/azure/cosmos-db/partitioning-overview" },
    { text: "Request Units and Pricing", url: "https://docs.microsoft.com/en-us/azure/cosmos-db/request-units" }
  ]}
/>
