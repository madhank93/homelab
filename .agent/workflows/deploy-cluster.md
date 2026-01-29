---
description: The single source of truth for deploying the Talos Proxmox cluster (Infra + Bootstrap + Apps).
---
# Workflow: Deploy Cluster

**Trigger:** `/deploy-cluster`

**Intent:** reliability provision and bootstrap the 7-node HA Cluster from zero to ready.

## Workflow Steps

### 1. Pre-Flight Check (`homelab-architect`)
- Ensure `.env` exists with `PROXMOX_PASSWORD`.
- Explain that `just deploy-proxmox` is the ONE command to run.
  - It handles `pulumi up` -> `pulumi refresh` (IP Lag) -> `pulumi up`.

### 2. Execution (Infrastructure)
- **User Action:** Run `just deploy-proxmox`.
- **Agent Monitoring:** Watch for:
  - "waiting for ip" (handled by script?)
  - HA Bootstrap race conditions (handled by Go code targeting `k8s-controller1`).
  - File generation: `./kubeconfig` and `./talosconfig`.

### 3. Verification (Apps)
- **User Action:** Run `./verify_cluster.sh`.
- **Agent Monitoring:**
  - **VIP:** Is `192.168.1.100` reachable?
  - **CNI:** Did Cilium install? (Nodes -> Ready?)
  - **Nginx:** Did `curl` return `Welcome to nginx!`?

### 4. Failure Handling
- If `deploy-proxmox` fails:
  - Run `just pulumi proxmox refresh` manually.
  - Check Proxmox UI for VM errors.
- If `verify_cluster.sh` fails:
  - `export KUBECONFIG=./kubeconfig`
  - `kubectl get nodes` to check API.
  - `talosctl health --nodes 192.168.1.203` (Controller 1 IP).
