+++
title = "CloudNativePG"
description = "CloudNativePG operator — Postgres lifecycle management for stateful workloads."
weight = 10
+++

## What is CloudNativePG?

[CloudNativePG](https://cloudnative-pg.io/) is a Kubernetes operator that manages the full lifecycle of PostgreSQL clusters — provisioning, configuration, backup, failover, and rolling updates — through a `Cluster` custom resource. It is the CNCF-recommended approach for running Postgres on Kubernetes.

## Why CloudNativePG?

Managing stateful Postgres directly with StatefulSets requires manual handling of replication, failover, connection pooling, and backup. CNPG encapsulates all of that into a single `Cluster` CR, keeps Postgres configuration as code, and integrates with Kubernetes storage (Longhorn PVCs) and monitoring (ServiceMonitor).

## How It's Used Here

CNPG runs as an operator in the `cnpg-system` namespace. It is used by n8n to provision a dedicated Postgres cluster — rather than bundling a Postgres sidecar in the n8n Helm chart, n8n's database is a first-class CNPG `Cluster` resource managed separately in the `n8n` namespace.

See [n8n](/workloads/automation/n8n/) for the Cluster CR definition and how n8n connects to it.

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Helm chart | `cloudnative-pg` v0.27.1 | Pinned version |
| Namespace | `cnpg-system` | Operator runs cluster-wide |
| Chart repo | `cloudnative-pg.github.io/charts` | Official CNPG chart repo |

Source: [`workloads/databases/cnpg.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/databases/cnpg.go)

## Troubleshooting

### Cluster Stuck in `Creating`

```bash
kubectl get cluster -n n8n
kubectl describe cluster n8n-db -n n8n
kubectl logs -n cnpg-system -l app.kubernetes.io/name=cloudnative-pg
```

### Pod Won't Schedule (PVC Pending)

```bash
kubectl get pvc -n n8n
kubectl describe pvc <pvc-name> -n n8n
```

Longhorn must be healthy and have sufficient free space. Check [Longhorn](/workloads/storage/longhorn/) status.
