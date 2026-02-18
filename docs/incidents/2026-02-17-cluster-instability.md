# Incident Report: Kubernetes Cluster Instability (2026-02-17)

## 1. Executive Summary
The Kubernetes cluster experienced a major outage characterized by API Server unavailability, intermittent node failures, and loss of control plane quorum. The root cause was identified as **severe resource contention on the underlying Proxmox host**, which starved the Control Plane nodes of CPU and Memory, causing critical system components (`etcd`, `containerd`, `kube-apiserver`) to time out and fail.

## 2. Root Cause Analysis (RCA)

### Primary Failure Mechanism: "Noisy Neighbor" Effect
The Proxmox host was overcommitted. High load from other VMs or processes on the host caused the Control Plane VMs (`k8s-controller-*`) to be starved of CPU cycles and memory access.

### Failure Chain of Events:
1.  **Resource Verification:** Proxmox host CPU/RAM usage spiked, reducing available resources for VMs.
2.  **Control Plane Starvation:** The Control Plane VMs, running with standard priority (`CpuUnits: 100`), could not get enough CPU time to process requests.
3.  **Etcd Latency & Timeout:** `etcd` (the cluster's database) is extremely sensitive to disk and network latency. Due to starvation, `etcd` heartbeats timed out, causing the cluster to lose quorum (leader election failed).
4.  **API Server Failure:**
    *   Without a healthy `etcd`, the `kube-apiserver` cannot function and becomes unresponsive.
    *   The `containerd` runtime on the nodes also began to fail health checks due to resource starvation, preventing it from even starting the API server container.
5.  **VIP Loss:** The Virtual IP (`192.168.1.100`), managed by `kube-vip`, stopped advertising because the underlying nodes were unhealthy, making the cluster unreachable via the standard endpoint.

## 3. Remediation Steps Taken

To recover the cluster and prevent recurrence, we prioritized the stability of the Control Plane over other workloads.

### 1. Resource Prioritization (The Fix)
We modified the infrastructure code (`infra/pulumi/proxmox.go` and `talos.go`) to implementstrict resource reservations:

*   **Control Plane Nodes (`k8s-controller-*`):**
    *   **CPU Priority:** Increased `CpuUnits` from **100** to **1024**. This ensures these VMs get ~10x more CPU time than standard VMs during contention.
    *   **Memory Guarantee:** Set `Balloon` to **4096** (matching the total RAM). This disables memory ballooning, guaranteeing the full 4GB of RAM is always physically backed and locked for these VMs.
*   **Worker Nodes:**
    *   **Resource Cap:** Reduced guaranteed memory and kept CPU priority at standard (`100`) to free up host resources for the Control Plane.

### 2. Operational Recovery
*   **Manual Intervention:** We bypassed the failing VIP (`192.168.1.100`) and accessed the cluster directly via the recovered node `192.168.1.171`.
*   **Service Restart:** We restarted the `kubelet` service on node `192.168.1.171`, which triggered the reconciliation of the static pods (`kube-apiserver`, `etcd`, `kube-vip`), effectively bringing the control plane back online.

## 4. Future Prevention & Recommendations

To avoid this in the future, we have already applied the **Resource Reservation** fix. However, additional measures are recommended:

### 1. Monitoring & Alerting (Critical)
*   **Proxmox Metrics:** Set up alerts for Proxmox Host CPU/RAM usage > 80%.
*   **Etcd Latency:** Monitor `etcd_disk_wal_fsync_duration_seconds`. Spikes here are early warning signs of starvation.

### 2. Capacity Planning
*   **Host Headroom:** Ensure the Proxmox host always has at least 15-20% unallocated CPU/RAM to handle spikes without starving critical VMs.
*   **Anti-Affinity:** If you add more Proxmox hosts, spread the Control Plane VMs across different physical hosts to survive single-host failures.

### 3. Automatic Recovery
*   **Watchdog Timers:** Talos Linux has built-in soft watchdog timers. Ensure these are enabled to automatically reboot nodes if they freeze completely (which seems to have happened to `.206`).

## 5. Current Status
*   **Cluster Health:** **Healthy**. All nodes are `Ready`.
*   **API Access:** **Restored**. The VIP `192.168.1.100` is functioning correctly.
*   **Pending Action:** The Nvidia GPU Operator deployment requires a configuration update (removing strict `nodeSelector`) to proceed.
