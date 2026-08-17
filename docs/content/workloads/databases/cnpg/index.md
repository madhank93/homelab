+++
title = "CloudNativePG"
description = "CloudNativePG operator — Postgres lifecycle management for stateful workloads."
weight = 10
+++

[CloudNativePG](https://cloudnative-pg.io/) runs as an operator in `cnpg-system`
and turns a Postgres cluster into a single `Cluster` custom resource — replication,
failover, backups and rolling updates included.

It exists here for one reason: **n8n's database**. Rather than accept the Postgres
subchart bundled with the n8n Helm chart, n8n's database is a CNPG `Cluster` in the
`n8n` namespace. That makes the database independently upgradable, gives it real
failover, and puts its credentials under the operator's control instead of in a
values file — CNPG generates the `n8n-pg-app` Secret itself, which is why the n8n
database password is the one secret **not** stored in OpenBao.

See [n8n](@/workloads/automation/n8n/index.md) for the Cluster CR and how n8n
connects to it.

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `cnpg-system` | Operator watches all namespaces from here |
| Chart | `cloudnative-pg` | Version in the [Software Inventory](@/architecture/software-inventory.md) |

Source: {{ src(path="workloads/databases/cnpg.go") }}

## Troubleshooting

### Cluster Stuck in `Creating`

```bash
kubectl get cluster -n n8n
kubectl describe cluster n8n-pg -n n8n
kubectl logs -n cnpg-system -l app.kubernetes.io/name=cloudnative-pg
```

### Pod Won't Schedule (PVC Pending)

```bash
kubectl get pvc -n n8n
kubectl describe pvc <pvc-name> -n n8n
```

Longhorn must be healthy and have sufficient free space. Check [Longhorn](/workloads/storage/longhorn/) status.
