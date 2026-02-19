# Network Configuration: Moving Talos to 192.168.2.x Subnet

## Overview
This document outlines the architectural changes made to transition the Talos Homelab cluster from DHCP (`192.168.1.x`) to static IP assignments in the `192.168.2.x` range using a `/23` subnet mask.

### Problem Statement
Previously, Talos Control Plane and Worker nodes acquired dynamic IPs from the AT&T Residential Gateway's DHCP server (`192.168.1.64` - `192.168.1.253`). This caused frequent cluster breakdowns when DHCP leases expired or nodes rebooted, leading `etcd` to lose quorum as node IPs changed unexpectedly.

### Solution: Static IPs with Subnet Expansion
To guarantee stability without needing an enterprise-grade router for VLANs, we expanded the network's subnet mask on the router from `/24` (`255.255.255.0`) to `/23` (`255.255.254.0`). 

A `/23` subnet mathematically merges the `192.168.1.x` and `192.168.2.x` blocks into a single continuous local network.
- **Network Range:** `192.168.1.1` through `192.168.2.254`
- **Gateway:** `192.168.1.254` (The AT&T Gateway) 
- **DHCP Clients (Phones, TVs):** Stay in `192.168.1.64` - `192.168.1.253`
- **Static Servers (Talos):** Assigned to `192.168.2.x`

Because AT&T Gateway only manages DHCP for the `.1.x` pool, the entire `.2.x` range is completely free from IP conflicts.

## Configuration Steps

### 1. AT&T Router Configuration (Manual Requirement)
You must physically log into your AT&T Gateway and update the Subnet Mask.
1. Navigate to your gateway at `http://192.168.1.254`
2. Go to **Settings -> LAN -> IPv4**
3. Change the **Subnet Mask** from `255.255.255.0` to **`255.255.254.0`**
4. Save and let the router briefly restart its network services.

### 2. Infrastructure as Code (Pulumi) Changes
The Pulumi definitions in `infra/pulumi/talos.go` were updated to statically assign `192.168.2.x` IP addresses to the VMs immediately upon bootstrap, bypassing DHCP.

`infra/pulumi/proxmox.go`:
- Updated `NodeConfig` struct to add the `IP string` field.

`infra/pulumi/talos.go`:
- **Nodes Array:** Defined explicit IPs for every VM (e.g., `k8s-controller1` = `192.168.2.11`, `k8s-worker1` = `192.168.2.21`, etc.)
- **VIP Configuration:** Moved the Talos API Virtual IP (`vipIP` and `clusterEndpoint`) from `192.168.1.100` to **`192.168.2.10`**.
- **Talos Machine Config Patches:** Adjusted `patchTalosConfig` to dynamically rewrite the cluster configuration via YAML parsing:
  - Disables DHCP (`dhcp: false`).
  - Assigns the static address `<IP>/23`.
  - Configures the default route `0.0.0.0/0` targeting `192.168.1.254`.
  - Sets DNS resolvers explicitly to `1.1.1.1` and `192.168.1.254`.

`infra/pulumi/cilium.go`:
- **K8s Service Host:** Updated Helm configurations to point `k8sServiceHost` to the new VIP: `192.168.2.10`.
- Cylium IP pool logic remains unhindered as it generates load balancer blocks dynamically mapped alongside the interface subnets.

## Recovery Protocol
Since changing the Control Plane IP ranges fundamentally destroys the previous `etcd` configurations, an in-place upgrade is impossible. To apply these configurations cleanly:
1. Destroy the current cluster VMs:
   ```bash
   just pulumi talos destroy
   ```
2. Reconstruct the cluster:
   ```bash
   just pulumi talos up
   ```
