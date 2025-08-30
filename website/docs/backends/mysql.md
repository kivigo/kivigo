---
sidebar_position: 10
title: MySQL
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="MySQL"
  description="MySQL is the world's most popular open source relational database. It provides robust, scalable, and reliable data storage with full SQL support and ACID compliance for mission-critical applications."
  category="Relational Database"
  packageName="backend/mysql"
  importPath="github.com/azrod/kivigo/backend/mysql"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "ACID Transactions", supported: true, description: "Full ACID compliance" },
    { name: "SQL Queries", supported: true, description: "Advanced querying with SQL" },
    { name: "Indexing", supported: true, description: "Primary, secondary, and composite indexes" },
    { name: "Replication", supported: true, description: "Master-slave and master-master replication" },
    { name: "Partitioning", supported: true, description: "Table partitioning for large datasets" }
  ]}
  dependencies={[
    "MySQL server (8.0 or later recommended)",
    "MySQL database and table created",
    "Valid connection credentials with appropriate permissions"
  ]}
  installationNotes="Requires a running MySQL server. You can use MySQL locally, MySQL in Docker, or cloud services like AWS RDS, Google Cloud SQL, or Azure Database for MySQL."
  configurationExample={`package main

import (
    "database/sql"
    "time"

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/mysql"
)

func main() {
    // Basic configuration
    opt := mysql.NewOptions()
    opt.DataSourceName = "user:password@tcp(localhost:3306)/kivigo?parseTime=true"
    opt.TableName = "keyvalue"
    opt.KeyColumn = "key_name"
    opt.ValueColumn = "value_data"

    // Advanced configuration with connection pool
    advOpt := mysql.NewOptions()
    advOpt.DataSourceName = "user:password@tcp(mysql-server:3306)/kivigo?charset=utf8mb4&parseTime=true&loc=Local"
    advOpt.TableName = "kvstore"
    advOpt.KeyColumn = "id"
    advOpt.ValueColumn = "data"
    advOpt.CreatedAtColumn = "created_at"
    advOpt.UpdatedAtColumn = "updated_at"
    
    // Optional: Configure connection pool
    advOpt.MaxOpenConns = 25
    advOpt.MaxIdleConns = 5
    advOpt.ConnMaxLifetime = 300 * time.Second
    advOpt.ConnMaxIdleTime = 60 * time.Second
    
    // Optional: Enable prepared statements
    advOpt.UsePreparedStatements = true
    
    // Create backend
    kvStore, err := mysql.New(advOpt)
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

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/mysql"
)

type OrderData struct {
    OrderID      string    \`json:"order_id"\`
    CustomerID   string    \`json:"customer_id"\`
    Items        []OrderItem \`json:"items"\`
    TotalAmount  float64   \`json:"total_amount"\`
    Status       string    \`json:"status"\`
    CreatedAt    time.Time \`json:"created_at"\`
    UpdatedAt    time.Time \`json:"updated_at"\`
}

type OrderItem struct {
    ProductID string  \`json:"product_id"\`
    Quantity  int     \`json:"quantity"\`
    Price     float64 \`json:"price"\`
}

func main() {
    // Setup
    opt := mysql.NewOptions()
    opt.DataSourceName = "kivigo_user:password@tcp(localhost:3306)/kivigo_db?parseTime=true"
    opt.TableName = "orders"
    opt.KeyColumn = "order_key"
    opt.ValueColumn = "order_data"

    kvStore, err := mysql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store order data
    order := OrderData{
        OrderID:    "ORD-2024-001",
        CustomerID: "CUST-456",
        Items: []OrderItem{
            {ProductID: "PROD-123", Quantity: 2, Price: 29.99},
            {ProductID: "PROD-456", Quantity: 1, Price: 89.99},
        },
        TotalAmount: 149.97,
        Status:      "pending",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    err = client.Set(ctx, "order:ORD-2024-001", order)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve order data
    var retrievedOrder OrderData
    err = client.Get(ctx, "order:ORD-2024-001", &retrievedOrder)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Order: %+v\\n", retrievedOrder)
    
    // Update order status
    retrievedOrder.Status = "processing"
    retrievedOrder.UpdatedAt = time.Now()
    
    err = client.Set(ctx, "order:ORD-2024-001", retrievedOrder)
    if err != nil {
        log.Fatal(err)
    }
    
    // List orders for customer
    keys, err := client.List(ctx, "order:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Order Keys: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/azrod/kivigo/backend/mysql"
)

func main() {
    opt := mysql.NewOptions()
    opt.DataSourceName = "user:password@tcp(localhost:3306)/kivigo?parseTime=true"
    opt.TableName = "keyvalue"

    kvStore, err := mysql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("MySQL unhealthy: %v", err)
        return
    }
    
    log.Println("MySQL is healthy")
    
    // Advanced health monitoring with connection pool stats
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            start := time.Now()
            
            // Test database connectivity and performance
            testKey := "health:check"
            testData := map[string]interface{}{
                "timestamp": time.Now().Unix(),
                "server":    "mysql-health-monitor",
                "check_id":  time.Now().Format("20060102-150405"),
            }
            
            // Test write operation
            if err := kvStore.SetRaw(ctx, testKey, []byte("health-check")); err != nil {
                log.Printf("Health check write failed: %v", err)
                continue
            }
            
            // Test read operation
            _, err := kvStore.GetRaw(ctx, testKey)
            if err != nil {
                log.Printf("Health check read failed: %v", err)
                continue
            }
            
            latency := time.Since(start)
            log.Printf("Health check passed (latency: %v)", latency)
            
            // Monitor for performance degradation
            if latency > 500*time.Millisecond {
                log.Printf("WARNING: High MySQL latency detected: %v", latency)
            }
            
            // Optional: Log connection pool statistics
            // (Implementation depends on your MySQL driver configuration)
            
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

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/mysql"
)

func main() {
    opt := mysql.NewOptions()
    opt.DataSourceName = "user:password@tcp(localhost:3306)/kivigo?parseTime=true"
    opt.TableName = "inventory"
    opt.KeyColumn = "item_key"
    opt.ValueColumn = "item_data"

    kvStore, err := mysql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch insert inventory data
    inventoryData := map[string]interface{}{
        "inventory:laptop:model-x1": map[string]interface{}{
            "sku": "LAP-X1-001", "quantity": 45, "price": 1299.99,
            "supplier": "TechCorp", "last_updated": time.Now().Format(time.RFC3339),
        },
        "inventory:phone:model-p1": map[string]interface{}{
            "sku": "PHN-P1-002", "quantity": 120, "price": 799.99,
            "supplier": "MobileTech", "last_updated": time.Now().Format(time.RFC3339),
        },
        "inventory:tablet:model-t1": map[string]interface{}{
            "sku": "TAB-T1-003", "quantity": 67, "price": 499.99,
            "supplier": "TabletInc", "last_updated": time.Now().Format(time.RFC3339),
        },
        "inventory:monitor:model-m1": map[string]interface{}{
            "sku": "MON-M1-004", "quantity": 89, "price": 329.99,
            "supplier": "DisplayCorp", "last_updated": time.Now().Format(time.RFC3339),
        },
    }
    
    err = client.BatchSet(ctx, inventoryData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch retrieve inventory for reporting
    keys := []string{
        "inventory:laptop:model-x1",
        "inventory:phone:model-p1",
        "inventory:tablet:model-t1",
        "inventory:monitor:model-m1",
    }
    
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    totalValue := 0.0
    for key, value := range results {
        fmt.Printf("%s: %v\\n", key, value)
        
        // Calculate total inventory value
        if item, ok := value.(map[string]interface{}); ok {
            if price, ok := item["price"].(float64); ok {
                if qty, ok := item["quantity"].(float64); ok {
                    totalValue += price * qty
                }
            }
        }
    }
    
    fmt.Printf("Total Inventory Value: $%.2f\\n", totalValue)
    
    // Batch remove discontinued items
    discontinuedKeys := []string{"inventory:old-model:discontinued"}
    err = client.BatchDelete(ctx, discontinuedKeys)
    if err != nil {
        log.Printf("Warning: Some discontinued items deletion failed: %v", err)
    }
    
    fmt.Println("Inventory batch operations completed")
}`}
  notes={[
    "Ensure proper indexing on key columns for optimal performance",
    "Use connection pooling to handle concurrent requests efficiently",
    "Consider using READ COMMITTED isolation level for most use cases",
    "MySQL 8.0 introduces better JSON support for storing complex values",
    "Monitor slow query log to identify performance bottlenecks",
    "Use prepared statements to prevent SQL injection and improve performance",
    "Consider partitioning large tables for better performance",
    "Regular backups are essential for data persistence and disaster recovery"
  ]}
  links={[
    { text: "MySQL Official Documentation", url: "https://dev.mysql.com/doc/" },
    { text: "Go MySQL Driver", url: "https://github.com/go-sql-driver/mysql" },
    { text: "MySQL Performance Tuning", url: "https://dev.mysql.com/doc/refman/8.0/en/optimization.html" },
    { text: "MySQL Cloud Solutions", url: "https://www.mysql.com/cloud/" }
  ]}
/>
