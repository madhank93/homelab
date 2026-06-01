---
description: Diagnose and resolve issues with the Talos Proxmox cluster using specific tools (talosctl, kubectl).
---
# Workflow: Troubleshoot Cluster

**Trigger:** `/troubleshoot-cluster`

**Intent:** Resolve common "Not Ready", "Connection Refused", or "Deployment Failed" states.

## Workflow Steps

### 1. Triage: The Layer Check
- **Layer 0 (Proxmox):**
  - Are VMs running?
  - Do they have IPs? (QEMU Agent)
- **Layer 1 (Talos API):**
  - Run `talosctl health --nodes <CONTROLLER_IP>`.
  - **Critical:** Do NOT trust the VIP (192.168.1.100) if the API is flaky. Use the node IP directly (e.g., `192.168.1.203`).
- **Layer 2 (Kubernetes API):**
  - `export KUBECONFIG=./kubeconfig`
  - `kubectl get nodes -o wide`
  - Are nodes responding? Are they `NotReady`? (CNI Issue?)

### 2. Common Fixes

#### Issue: "Waiting for IP address" (Pulumi)
- **Cause:** DHCP Lag / QEMU Agent delay.
- **Fix:** `just pulumi proxmox refresh` then `just pulumi proxmox up`.

#### Issue: "Nodes NotReady"
- **Cause:** CNI (Cilium) not installed or configured wrong.
- **Fix:** Run `./verify_cluster.sh` (installs CNI).
- **Check:** `kubectl get pods -n kube-system` (Look for `Init:RunContainerError` -> means Helm vals are wrong).

#### Issue: "etcd quorum lost" / "connection refused"
- **Cause:** Bootstrap targeted the VIP or wrong node.
- **Fix:**
  - Verify `proxmox.go` targets `k8s-controller1` specifically.
  - Reboot control plane nodes one by one.

### 3. Essential Commands
- **Logs:** `talosctl logs -n <NODE_IP> kubelet`
- **Services:** `talosctl services -n <NODE_IP>`
- **Reset:** `just pulumi proxmox destroy` (Nuke it).
