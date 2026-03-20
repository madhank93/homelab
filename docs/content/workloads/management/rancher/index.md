+++
title = "Rancher"
description = "Multi-cluster Kubernetes management UI with OpenBao secrets integration."
weight = 10
+++

## What is Rancher?

[Rancher](https://rancher.com/) is an open-source Kubernetes management platform that provides a web UI for managing multiple clusters, workloads, nodes, storage, and networking. It includes Fleet for GitOps-based multi-cluster management and a Helm app catalog.

## Why Rancher?

Rancher provides deeper cluster management capabilities than Headlamp — including node management, cluster import/creation, Helm app deployment, and the Fleet GitOps system. Both Rancher and Headlamp run in this homelab; Headlamp for quick Kubernetes API browsing, Rancher for cluster administration and fleet management.

## How It's Used Here

Rancher runs in the `cattle-system` namespace with 3 replicas (for HA within the cluster). Bootstrap password is managed by OpenBao via a dedicated `secret-sync` Deployment (Pattern B).

Source: [`workloads/management/rancher.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/management/rancher.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `cattle-system` | Rancher convention |
| HTTPRoute | `rancher.madhan.app` → `rancher:80` | Gateway API |
| `hostname` | `rancher.madhan.app` | Used for redirect URLs |
| Replicas | `3` | HA within cluster |
| `agentTLSMode` | `system-store` | Use system CA store for agent TLS |
| `auditLog.level` | `0` | Audit logging disabled (set to 1+ for production) |
| `ingress.enabled` | `false` | Routing via Gateway API |
| `existingBootstrapPassword` | `rancher-bootstrap` | CSI-synced Secret from OpenBao |
| Resources (limits) | `1000m` / `2Gi` | 3 replicas each |
| `--kube-version` | `1.30.0` | Required for chart compatibility with this k8s version |

## Secrets (OpenBao)

Pattern B (secretObjects sync). `BOOTSTRAP_PASSWORD` is fetched from OpenBao (`secret/data/rancher`) and synced into the `rancher-bootstrap` k8s Secret.

**Why a dedicated secret-sync Deployment?** Rancher's Helm chart does not support `extraVolumes`. A `secret-sync` Deployment with a `pause` container mounts the CSI volume to trigger secretObjects sync.

```go
"existingBootstrapPassword":  "rancher-bootstrap",
"bootstrapPasswordSecretKey": "BOOTSTRAP_PASSWORD",
```

## How It Connects

```
Browser → rancher.madhan.app
  → homelab-gateway → rancher:80
  → Rancher pod(s) in cattle-system
  → Kubernetes API (reads all cluster resources)
  → OpenBao CSI → secret-sync pod → rancher-bootstrap k8s Secret
```

## Screenshots

![Rancher cluster dashboard showing workload health, node status, and resource usage](/assets/screenshots/rancher/cluster-dashboard.png)

## Troubleshooting

### Bootstrap Password Not Working

**Symptoms:** Rancher login shows "Invalid credentials" for the bootstrap admin user.

**Diagnosis:**

```bash
# Check secret-sync pod is running
kubectl get pods -n cattle-system -l app=secret-sync
kubectl logs -n cattle-system -l app=secret-sync

# Verify bootstrap secret is populated
kubectl get secret rancher-bootstrap -n cattle-system -o yaml

# Verify OpenBao has the password
kubectl exec -n openbao openbao-0 -- bao kv get secret/rancher
```

### Rancher Agent Connectivity

If Rancher reports cluster agents as disconnected after a restart:

```bash
# Check Rancher pods
kubectl get pods -n cattle-system

# Check cattle-cluster-agent
kubectl get pods -n cattle-system -l app=cattle-cluster-agent
```

### Helm Flag `--kube-version`

Rancher's Helm chart requires explicit `--kube-version` for chart validation when the Kubernetes version is newer than what the chart explicitly supports. This is set in the CDK8s Helm properties via `HelmFlags`.
