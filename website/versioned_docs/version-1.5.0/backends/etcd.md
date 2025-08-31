---
sidebar_position: 5
title: etcd
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="etcd"
  description="etcd is a distributed, reliable key-value store for the most critical data of a distributed system. It's the backing store for Kubernetes and provides strong consistency guarantees."
  category="Distributed Store"
  packageName="backend/etcd"
  importPath="github.com/azrod/kivigo/backend/etcd"
  features={[
    { name: "Basic Operations", supported: true },
    { name: "Batch Operations", supported: true },
    { name: "Health Checks", supported: true },
    { name: "Persistence", supported: true },
    { name: "MVCC", supported: true },
    { name: "Transactions", supported: true },
    { name: "Watches", supported: false, description: "Not exposed through KiviGo interface" },
    { name: "Leases", supported: false, description: "Not exposed through KiviGo interface" }
  ]}
  dependencies={[
    "etcd server (version 3.4+)",
    "Network connectivity to etcd cluster"
  ]}
  configurationExample={`// etcd configuration example
opt := etcd.DefaultOptions()
opt.Endpoints = []string{"localhost:2379"}
kvStore, err := etcd.New(opt)`}
  usageExample={`// Basic etcd usage
err = client.Set(ctx, "cluster/config", configData)
var config ConfigType
err = client.Get(ctx, "cluster/config", &config)`}
  notes={[
    "etcd provides linearizable reads and writes",
    "Uses Raft consensus for distributed coordination",
    "Optimized for metadata and configuration storage",
    "Default Kubernetes backing store"
  ]}
  links={[
    { text: "etcd Documentation", url: "https://etcd.io/docs/" }
  ]}
/>
