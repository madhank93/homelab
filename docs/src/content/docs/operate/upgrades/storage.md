---
title: Longhorn and CloudNativePG
description: Upgrade Longhorn one minor at a time and the CNPG operator, keeping every volume healthy.
sidebar:
  order: 15
---

Phase 6. Risk: medium, storage disruption is possible. Stateful workloads depend on both.

## Longhorn

Pin: `helm:https://charts.longhorn.io/longhorn@X.Y.Z` in [`workloads/cdk8s.yaml`](src:workloads/cdk8s.yaml), then [regenerate imports](/operate/upgrades/cdk8s-imports/).

:::danger[One minor at a time]
1.10 → 1.11 → 1.12, never 1.10 → 1.12. Longhorn supports upgrades only from the previous minor.
:::

Before syncing: every volume `healthy`, none degraded or faulted. No V2 data-engine volumes in use.

```bash frame="terminal"
just synth && git push
kubectl get nodes.longhorn.io -n longhorn-system
kubectl get volumes.longhorn.io -n longhorn-system \
  -o custom-columns=NAME:.metadata.name,STATE:.status.state,ROBUST:.status.robustness
kubectl get engineimages -n longhorn-system   # new image deployed on every node
```

Volume engines upgrade live, one per node at a time (`concurrentAutomaticEngineUpgradePerNodeLimit: 1` in [`longhorn.go`](src:workloads/storage/longhorn.go)). A strict-local volume such as `ollama` needs a detach before its engine moves.

## CloudNativePG

Pins in [`config/versions.yaml`](src:config/versions.yaml): `charts.cloudnative-pg` (operator) and `images.n8n-postgres` (the `n8n-pg` operand). Operator upgrades are backwards-compatible with existing `Cluster` CRs. An operand bump within one PostgreSQL major is a rolling restart; a major bump is not.

```bash frame="terminal"
kubectl get cluster -A   # every cluster healthy
```

`n8n-pg` is a single instance with no scheduled backup. Take a `pg_dump` first — see [n8n](/operate/upgrades/workloads/).
