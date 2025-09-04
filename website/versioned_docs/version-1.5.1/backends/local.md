---
sidebar_position: 6
title: BoltDB (Local)
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="BoltDB (Local)"
  description="BoltDB is a pure Go key/value store inspired by Howard Chu's LMDB. It provides a simple, fast, and reliable database for projects that don't require a full database server."
  category="Embedded Store"
  packageName="backend/local"
  importPath="github.com/azrod/kivigo/backend/local"
  features={[
    { name: "Basic Operations", supported: true },
    { name: "Batch Operations", supported: true },
    { name: "Health Checks", supported: true },
    { name: "Persistence", supported: true },
    { name: "ACID Transactions", supported: true },
    { name: "Single Writer", supported: true },
    { name: "Memory Mapped", supported: true },
    { name: "Cross Platform", supported: true }
  ]}
  dependencies={[
    "No external services required",
    "File system write permissions"
  ]}
  configurationExample={`// BoltDB configuration
opt := local.Option{Path: "./data.db"}
kvStore, err := local.New(opt)`}
  usageExample={`// Basic BoltDB usage  
err = client.Set(ctx, "config", configData)
var config ConfigType
err = client.Get(ctx, "config", &config)`}
  notes={[
    "Single writer, multiple readers",
    "Zero-configuration embedded database",
    "ACID compliant transactions",
    "Perfect for single-node applications"
  ]}
  links={[
    { text: "BoltDB Documentation", url: "https://github.com/etcd-io/bbolt" }
  ]}
/>
