---
sidebar_position: 4
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="Consul"
  description="Consul is a service discovery and configuration management tool. It provides a distributed key-value store that's perfect for configuration management and service coordination in microservices architectures."
  category="Distributed Store"
  packageName="backend/consul"
  importPath="github.com/azrod/kivigo/backend/consul"
  features={[
    { name: "Basic Operations", supported: true },
    { name: "Batch Operations", supported: true },
    { name: "Health Checks", supported: true },
    { name: "Persistence", supported: true },
    { name: "Multi-DC Support", supported: true },
    { name: "ACLs", supported: true },
    { name: "Watches", supported: false, description: "Not exposed through KiviGo interface" },
    { name: "Service Discovery", supported: false, description: "Not exposed through KiviGo interface" }
  ]}
  dependencies={[
    "Consul server (version 1.8+)",
    "Network connectivity to Consul cluster"
  ]}
  installationNotes="Requires a running Consul server or cluster. Consul provides strong consistency guarantees through the Raft consensus algorithm."
  configurationExample={`package main

import (
    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/consul"
)

func main() {
    // Basic configuration
    opt := consul.DefaultOptions()
    opt.Address = "localhost:8500"

    // Custom configuration with auth
    customOpt := &consul.Config{
        Address: "consul.example.com:8500",
        Token:   "your-acl-token",
        Scheme:  "https",
        Datacenter: "dc1",
    }
    
    // Create backend
    kvStore, err := consul.New(customOpt)
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

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/consul"
)

type ServiceConfig struct {
    Name     string \`json:"name"\`
    Port     int    \`json:"port"\`
    Replicas int    \`json:"replicas"\`
    Enabled  bool   \`json:"enabled"\`
}

func main() {
    // Setup
    opt := consul.DefaultOptions()
    opt.Address = "localhost:8500"

    kvStore, err := consul.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store service configuration
    config := ServiceConfig{
        Name:     "auth-service",
        Port:     8080,
        Replicas: 3,
        Enabled:  true,
    }
    
    err = client.Set(ctx, "services/auth-service/config", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve configuration
    var retrievedConfig ServiceConfig
    err = client.Get(ctx, "services/auth-service/config", &retrievedConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Service Config: %+v\\n", retrievedConfig)
    
    // List all service configurations
    keys, err := client.List(ctx, "services/")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Service keys: %v\\n", keys)
}`}
  healthCheckExample={`// Consul health check example
err := kvStore.Health(ctx)
if err != nil {
    log.Printf("Consul unhealthy: %v", err)
} else {
    log.Println("Consul is healthy")
}`}
  notes={[
    "Consul provides strong consistency through Raft consensus",
    "All writes go through the leader, reads can be eventually consistent",
    "Use ACL tokens for production security",
    "Consul supports multi-datacenter deployments",
    "Consider using Consul Connect for service mesh capabilities"
  ]}
  links={[
    { text: "Consul Documentation", url: "https://www.consul.io/docs" },
    { text: "Consul KV Store Guide", url: "https://www.consul.io/docs/dynamic-app-config/kv" }
  ]}
/>
