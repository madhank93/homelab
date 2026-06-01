+++
title = "Hardware"
description = "Physical hardware overview: cluster nodes, networking, and GPU."
weight = 20
sort_by = "weight"
+++

This section covers the physical hardware running the homelab Kubernetes cluster.

## Cluster Overview

The cluster runs on 7 virtual machines provisioned by Pulumi on a single Proxmox host:

| Role | Count | Hostnames | IPs |
|------|-------|-----------|-----|
| Control Plane | 3 | k8s-controller1–3 | 192.168.1.211–213 |
| Worker | 3 | k8s-worker1–3 | 192.168.1.221–223 |
| GPU Worker | 1 | k8s-worker4 | 192.168.1.224 |

**VIP (API server):** `192.168.1.210:6443`
**Load Balancer pool:** `192.168.1.220–230` (Cilium L2 announcements)

## Key Specs Per VM

| Node | vCPUs | RAM | Disk | Special |
|------|-------|-----|------|---------|
| k8s-controller1–3 | 4 | 6 GiB | 30 GiB | VIP on eth0 |
| k8s-worker1–3 | 4 | 6 GiB | 125 GiB | Longhorn storage |
| k8s-worker4 | 4 | 6 GiB | 125 GiB | NVIDIA RTX 5070 Ti GPU (PCIe passthrough) |
