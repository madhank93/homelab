---
title: CloudNativePG
description: Postgres operator that runs n8n's database and owns its credentials.
sidebar:
  order: 10
---

The CloudNativePG operator turns a `Cluster` resource into a running Postgres. It runs one database today: n8n's. [Upstream docs](https://cloudnative-pg.io/documentation/current/).

| | |
|---|---|
| Namespace | `cnpg-system` (operator); clusters live in each app's namespace |
| Source | [`cnpg.go`](src:workloads/databases/cnpg.go) |
| Exposure | — |
| Storage | Per cluster; n8n's is 10Gi on `longhorn` |
| Auth | — |

Why an operator rather than the chart's bundled Postgres: [CloudNativePG for app databases](/reference/decisions/cnpg/).

## How it fits

For a `Cluster` named `<name>`, the operator creates the Secret `<name>-app` (app user's credentials) and the Service `<name>-rw` (primary). Apps read the password from that Secret, so it is the one credential not stored in OpenBao. The [n8n](/workloads/automation/n8n/) page has the `Cluster` spec.

The chart runs with default values.

## Operate

The operand image is set per cluster with `imageName`. A minor bump within one major is an in-place rolling update; a major bump needs a new cluster or `pg_upgrade`.

## Troubleshooting

### Cluster stuck in `Creating`

```bash frame="terminal"
kubectl describe cluster n8n-pg -n n8n
kubectl logs -n cnpg-system -l app.kubernetes.io/name=cloudnative-pg --tail=50
```

### Pod pending on its PVC

Longhorn cannot place the volume: `kubectl describe pvc -n n8n`, then check [Longhorn](/workloads/storage/longhorn/).
