---
sidebar_position: 11
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="PostgreSQL"
  description="PostgreSQL is a powerful, open source object-relational database system with a strong reputation for reliability, feature robustness, and performance. It supports both SQL and JSON querying."
  category="Relational Database"
  packageName="backend/postgresql"
  importPath="github.com/azrod/kivigo/backend/postgresql"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "ACID Transactions", supported: true, description: "Full ACID compliance with advanced isolation levels" },
    { name: "JSON Support", supported: true, description: "Native JSON and JSONB data types" },
    { name: "Advanced Indexing", supported: true, description: "GiST, GIN, BRIN, and custom index types" },
    { name: "Streaming Replication", supported: true, description: "Hot standby and streaming replication" },
    { name: "Extensions", supported: true, description: "Rich ecosystem of extensions (PostGIS, pg_cron, etc.)" }
  ]}
  dependencies={[
    "PostgreSQL server (12 or later recommended)",
    "PostgreSQL database and table created",
    "Valid connection credentials with appropriate permissions"
  ]}
  installationNotes="Requires a running PostgreSQL server. You can use PostgreSQL locally, in Docker, or cloud services like AWS RDS, Google Cloud SQL, or Azure Database for PostgreSQL."
  configurationExample={`package main

import (
    "time"

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/postgresql"
)

func main() {
    // Basic configuration
    opt := postgresql.NewOptions()
    opt.DataSourceName = "postgres://user:password@localhost:5432/kivigo?sslmode=disable"
    opt.TableName = "keyvalue"
    opt.KeyColumn = "key"
    opt.ValueColumn = "value"

    // Advanced configuration with SSL and connection pool
    advOpt := postgresql.NewOptions()
    advOpt.DataSourceName = "postgres://user:password@postgres-server:5432/kivigo?sslmode=require&sslcert=client-cert.pem&sslkey=client-key.pem&sslrootcert=ca-cert.pem"
    advOpt.TableName = "kv_store"
    advOpt.KeyColumn = "id"
    advOpt.ValueColumn = "data"
    advOpt.CreatedAtColumn = "created_at"
    advOpt.UpdatedAtColumn = "updated_at"
    
    // Optional: Configure connection pool
    advOpt.MaxOpenConns = 30
    advOpt.MaxIdleConns = 10
    advOpt.ConnMaxLifetime = 300 * time.Second
    advOpt.ConnMaxIdleTime = 90 * time.Second
    
    // Optional: Use JSONB for value storage
    advOpt.UseJSONB = true
    advOpt.EnableWAL = true
    
    // Create backend
    kvStore, err := postgresql.New(advOpt)
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
    "github.com/azrod/kivigo/backend/postgresql"
)

type EventLog struct {
    EventID     string    \`json:"event_id"\`
    UserID      string    \`json:"user_id"\`
    EventType   string    \`json:"event_type"\`
    Timestamp   time.Time \`json:"timestamp"\`
    Properties  map[string]interface{} \`json:"properties"\`
    UserAgent   string    \`json:"user_agent"\`
    IPAddress   string    \`json:"ip_address"\`
    SessionID   string    \`json:"session_id"\`
}

func main() {
    // Setup
    opt := postgresql.NewOptions()
    opt.DataSourceName = "postgres://kivigo_user:password@localhost:5432/kivigo_events?sslmode=disable"
    opt.TableName = "event_logs"
    opt.KeyColumn = "log_key"
    opt.ValueColumn = "log_data"
    opt.UseJSONB = true  // Use JSONB for better JSON performance

    kvStore, err := postgresql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store event log
    event := EventLog{
        EventID:   "evt-2024-001",
        UserID:    "user-789",
        EventType: "page_view",
        Timestamp: time.Now(),
        Properties: map[string]interface{}{
            "page_url":     "/docs/postgresql",
            "referrer":     "https://google.com",
            "duration_ms":  2345,
            "scroll_depth": 0.75,
        },
        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        IPAddress: "192.168.1.100",
        SessionID: "sess-abc123",
    }
    
    err = client.Set(ctx, "event:evt-2024-001", event)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve event log
    var retrievedEvent EventLog
    err = client.Get(ctx, "event:evt-2024-001", &retrievedEvent)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Event: %+v\\n", retrievedEvent)
    
    // Store user action event
    actionEvent := EventLog{
        EventID:   "evt-2024-002",
        UserID:    "user-789",
        EventType: "button_click",
        Timestamp: time.Now(),
        Properties: map[string]interface{}{
            "button_id":   "download-btn",
            "button_text": "Download Documentation",
            "page_url":    "/docs/postgresql",
            "coordinates": map[string]int{"x": 350, "y": 180},
        },
        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        IPAddress: "192.168.1.100",
        SessionID: "sess-abc123",
    }
    
    err = client.Set(ctx, "event:evt-2024-002", actionEvent)
    if err != nil {
        log.Fatal(err)
    }
    
    // List events for analysis
    keys, err := client.List(ctx, "event:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Event Keys: %v\\n", keys)
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/azrod/kivigo/backend/postgresql"
)

func main() {
    opt := postgresql.NewOptions()
    opt.DataSourceName = "postgres://user:password@localhost:5432/kivigo?sslmode=disable"
    opt.TableName = "keyvalue"

    kvStore, err := postgresql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("PostgreSQL unhealthy: %v", err)
        return
    }
    
    log.Println("PostgreSQL is healthy")
    
    // Advanced health monitoring with replication lag check
    ticker := time.NewTicker(45 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            start := time.Now()
            
            // Test write/read operations
            testKey := "health:check"
            testValue := map[string]interface{}{
                "timestamp":     time.Now().Unix(),
                "server":        "postgres-health-monitor",
                "check_id":      time.Now().Format("20060102-150405"),
                "random_data":   []int{1, 2, 3, 4, 5},
            }
            
            // Test JSONB write performance
            if err := kvStore.SetRaw(ctx, testKey, []byte("health-check")); err != nil {
                log.Printf("Health check write failed: %v", err)
                continue
            }
            
            // Test read performance
            _, err := kvStore.GetRaw(ctx, testKey)
            if err != nil {
                log.Printf("Health check read failed: %v", err)
                continue
            }
            
            latency := time.Since(start)
            log.Printf("Health check passed (latency: %v)", latency)
            
            // Monitor for performance issues
            if latency > 200*time.Millisecond {
                log.Printf("WARNING: High PostgreSQL latency: %v", latency)
            }
            
            // Optional: Check connection pool stats
            // (Available through database/sql.DBStats if using the underlying connection)
            
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
    "github.com/azrod/kivigo/backend/postgresql"
)

func main() {
    opt := postgresql.NewOptions()
    opt.DataSourceName = "postgres://user:password@localhost:5432/kivigo?sslmode=disable"
    opt.TableName = "time_series"
    opt.KeyColumn = "metric_key"
    opt.ValueColumn = "metric_data"
    opt.UseJSONB = true

    kvStore, err := postgresql.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch insert time series data
    timestamp := time.Now()
    timeSeriesData := map[string]interface{}{
        "metrics:cpu:server1:" + timestamp.Format("2006-01-02T15:04"): map[string]interface{}{
            "timestamp": timestamp.Unix(),
            "value":     78.5,
            "unit":      "percent",
            "tags":      map[string]string{"server": "web-01", "datacenter": "us-east-1"},
            "metadata":  map[string]interface{}{"cores": 8, "frequency": "2.4GHz"},
        },
        "metrics:memory:server1:" + timestamp.Format("2006-01-02T15:04"): map[string]interface{}{
            "timestamp": timestamp.Unix(),
            "value":     6.2,
            "unit":      "GB",
            "tags":      map[string]string{"server": "web-01", "datacenter": "us-east-1"},
            "metadata":  map[string]interface{}{"total_gb": 16, "type": "DDR4"},
        },
        "metrics:disk:server1:" + timestamp.Format("2006-01-02T15:04"): map[string]interface{}{
            "timestamp": timestamp.Unix(),
            "value":     245.8,
            "unit":      "GB",
            "tags":      map[string]string{"server": "web-01", "datacenter": "us-east-1"},
            "metadata":  map[string]interface{}{"total_gb": 500, "type": "SSD"},
        },
        "metrics:network:server1:" + timestamp.Format("2006-01-02T15:04"): map[string]interface{}{
            "timestamp": timestamp.Unix(),
            "value":     1024.5,
            "unit":      "Mbps",
            "tags":      map[string]string{"server": "web-01", "datacenter": "us-east-1"},
            "metadata":  map[string]interface{}{"interface": "eth0", "speed": "1Gbps"},
        },
    }
    
    err = client.BatchSet(ctx, timeSeriesData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch retrieve metrics for dashboard
    keys := make([]string, 0)
    for key := range timeSeriesData {
        keys = append(keys, key)
    }
    
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process metrics for dashboard display
    dashboardData := make(map[string]interface{})
    for key, value := range results {
        if metric, ok := value.(map[string]interface{}); ok {
            metricType := "unknown"
            if len(key) > 8 && key[:8] == "metrics:" {
                parts := key[8:]
                if colonIndex := 0; colonIndex < len(parts) {
                    for i, char := range parts {
                        if char == ':' {
                            metricType = parts[:i]
                            break
                        }
                    }
                }
            }
            
            dashboardData[metricType] = map[string]interface{}{
                "current_value": metric["value"],
                "unit":         metric["unit"],
                "last_updated": metric["timestamp"],
            }
        }
        
        fmt.Printf("%s: %v\\n", key, value)
    }
    
    fmt.Printf("Dashboard Data: %+v\\n", dashboardData)
    
    // Batch cleanup old metrics (older than 24 hours)
    cutoffTime := time.Now().Add(-24 * time.Hour)
    oldKeys := []string{
        "metrics:cpu:server1:" + cutoffTime.Format("2006-01-02T15:04"),
        "metrics:memory:server1:" + cutoffTime.Format("2006-01-02T15:04"),
    }
    
    err = client.BatchDelete(ctx, oldKeys)
    if err != nil {
        log.Printf("Warning: Old metrics cleanup failed: %v", err)
    }
    
    fmt.Println("Time series batch operations completed")
}`}
  notes={[
    "PostgreSQL JSONB provides better performance than JSON for complex queries",
    "Use proper indexes on JSONB columns for optimal query performance",
    "Connection pooling is crucial for high-concurrency applications",
    "Consider using table partitioning for time-series or large datasets",
    "PostgreSQL supports advanced features like window functions and CTEs",
    "Use EXPLAIN ANALYZE to optimize query performance",
    "WAL (Write-Ahead Logging) provides durability and enables point-in-time recovery",
    "PostgreSQL extensions like PostGIS add powerful geospatial capabilities"
  ]}
  links={[
    { text: "PostgreSQL Official Documentation", url: "https://www.postgresql.org/docs/" },
    { text: "Go PostgreSQL Driver (pq)", url: "https://github.com/lib/pq" },
    { text: "PostgreSQL JSON Functions", url: "https://www.postgresql.org/docs/current/functions-json.html" },
    { text: "PostgreSQL Performance Tuning", url: "https://wiki.postgresql.org/wiki/Performance_Optimization" }
  ]}
/>
