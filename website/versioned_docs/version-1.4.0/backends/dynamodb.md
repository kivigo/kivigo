---
sidebar_position: 13
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="Amazon DynamoDB"
  description="Amazon DynamoDB is a fully managed NoSQL database service that provides fast and predictable performance with seamless scalability. It's designed for applications that need consistent, single-digit millisecond latency at any scale."
  category="Cloud Database"
  packageName="backend/dynamodb"
  importPath="github.com/kivigo/kivigo/backend/dynamodb"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "Auto Scaling", supported: true, description: "Automatic capacity scaling" },
    { name: "Global Tables", supported: true, description: "Multi-region replication" },
    { name: "Point-in-Time Recovery", supported: true, description: "Continuous backups" },
    { name: "On-Demand Billing", supported: true, description: "Pay-per-request pricing" },
    { name: "DAX Caching", supported: true, description: "Microsecond latency with DynamoDB Accelerator" }
  ]}
  dependencies={[
    "AWS account with DynamoDB permissions",
    "AWS credentials configured (IAM role, access key, or AWS CLI)",
    "DynamoDB table created in your AWS region"
  ]}
  installationNotes="Requires AWS credentials and a pre-created DynamoDB table. You can use AWS Free Tier which includes 25 GB of storage and 25 WCU/RCU."
  configurationExample={`package main

import (
    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/backend/dynamodb"
    "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
    // Load AWS configuration
    awsConfig, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        panic(err)
    }

    // Basic configuration
    opt := dynamodb.NewOptions()
    opt.Config = awsConfig
    opt.TableName = "kivigo-keyvalue"
    opt.Region = "us-east-1"
    
    // Optional: Custom configuration
    opt.KeyAttribute = "id"        // Primary key attribute name
    opt.ValueAttribute = "data"    // Value attribute name
    opt.TTLAttribute = "expires"   // TTL attribute name (optional)
    
    // Optional: Configure read/write capacity for provisioned tables
    opt.ReadCapacityUnits = 5
    opt.WriteCapacityUnits = 5
    
    // Create backend
    kvStore, err := dynamodb.New(opt)
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
    "github.com/kivigo/kivigo/backend/dynamodb"
    "github.com/aws/aws-sdk-go-v2/config"
)

type SessionData struct {
    UserID    string \`json:"user_id"\`
    SessionID string \`json:"session_id"\`
    CreatedAt string \`json:"created_at"\`
    Data      map[string]interface{} \`json:"data"\`
}

func main() {
    // Setup
    awsConfig, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    opt := dynamodb.NewOptions()
    opt.Config = awsConfig
    opt.TableName = "kivigo-sessions"
    opt.Region = "us-east-1"
    
    kvStore, err := dynamodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store session data
    session := SessionData{
        UserID:    "user-456",
        SessionID: "sess-789",
        CreatedAt: "2024-01-01T00:00:00Z",
        Data: map[string]interface{}{
            "role":        "admin",
            "permissions": []string{"read", "write", "delete"},
            "theme":       "dark",
        },
    }
    
    err = client.Set(ctx, "session:sess-789", session)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve session data
    var retrievedSession SessionData
    err = client.Get(ctx, "session:sess-789", &retrievedSession)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Session: %+v\\n", retrievedSession)
    
    // List sessions for a user
    keys, err := client.List(ctx, "session:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Active Sessions: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/kivigo/kivigo/backend/dynamodb"
    "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
    awsConfig, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    opt := dynamodb.NewOptions()
    opt.Config = awsConfig
    opt.TableName = "kivigo-keyvalue"
    opt.Region = "us-east-1"
    
    kvStore, err := dynamodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("DynamoDB unhealthy: %v", err)
        return
    }
    
    log.Println("DynamoDB is healthy")
    
    // Advanced health monitoring with metrics
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            start := time.Now()
            if err := kvStore.Health(ctx); err != nil {
                log.Printf("Health check failed: %v", err)
            } else {
                latency := time.Since(start)
                log.Printf("Health check passed (latency: %v)", latency)
                
                // Alert if latency is high
                if latency > 100*time.Millisecond {
                    log.Printf("WARNING: High latency detected: %v", latency)
                }
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
    "github.com/kivigo/kivigo/backend/dynamodb"
    "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
    awsConfig, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    opt := dynamodb.NewOptions()
    opt.Config = awsConfig
    opt.TableName = "kivigo-keyvalue"
    opt.Region = "us-east-1"
    
    kvStore, err := dynamodb.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch write - up to 25 items per batch request
    batchData := map[string]interface{}{
        "metric:cpu:server1":    map[string]interface{}{"value": 85.5, "timestamp": "2024-01-01T12:00:00Z"},
        "metric:memory:server1": map[string]interface{}{"value": 67.2, "timestamp": "2024-01-01T12:00:00Z"},
        "metric:disk:server1":   map[string]interface{}{"value": 45.8, "timestamp": "2024-01-01T12:00:00Z"},
        "metric:cpu:server2":    map[string]interface{}{"value": 92.1, "timestamp": "2024-01-01T12:00:00Z"},
        "metric:memory:server2": map[string]interface{}{"value": 78.9, "timestamp": "2024-01-01T12:00:00Z"},
    }
    
    err = client.BatchSet(ctx, batchData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch read - up to 100 items per batch request
    keys := []string{
        "metric:cpu:server1",
        "metric:memory:server1", 
        "metric:disk:server1",
        "metric:cpu:server2",
        "metric:memory:server2",
    }
    
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
    "DynamoDB has a maximum item size of 400 KB",
    "Batch operations can process up to 25 items (write) or 100 items (read) per request",
    "Choose your partition key carefully to avoid hot partitions",
    "Use on-demand billing for unpredictable workloads, provisioned for predictable ones",
    "Global Secondary Indexes (GSI) allow querying on non-key attributes",
    "DynamoDB Streams can trigger Lambda functions on data changes",
    "Consider using DynamoDB Accelerator (DAX) for microsecond latency requirements",
    "The free tier includes 25 GB storage, 25 WCU, and 25 RCU per month"
  ]}
  links={[
    { text: "Amazon DynamoDB Documentation", url: "https://docs.aws.amazon.com/dynamodb/" },
    { text: "AWS SDK for Go v2", url: "https://github.com/aws/aws-sdk-go-v2" },
    { text: "DynamoDB Best Practices", url: "https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html" },
    { text: "DynamoDB Pricing", url: "https://aws.amazon.com/dynamodb/pricing/" }
  ]}
/>
