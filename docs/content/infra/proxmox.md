+++
title = "Proxmox Cluster"
weight = 20
+++

# Proxmox Kubernetes Cluster

The local Kubernetes cluster is hosted on Proxmox VE. The infrastructure is defined in `infra/pulumi/proxmox.go`.

## Node Configuration

The cluster consists of 7 nodes: 3 Control Plane nodes and 4 Worker nodes.

| Node Name | Role | Cores | Memory (MB) | Disk (GB) | Features |
|-----------|------|-------|-------------|-----------|----------|
| `k8s-controller1` | Control Plane | 2 | 4096 | 30 | |
| `k8s-controller2` | Control Plane | 2 | 4096 | 30 | |
| `k8s-controller3` | Control Plane | 2 | 4096 | 30 | |
| `k8s-worker1` | Worker | 4 | 8192 | 125 | |
| `k8s-worker2` | Worker | 4 | 8192 | 125 | |
| `k8s-worker3` | Worker | 4 | 8192 | 125 | |
| `k8s-worker4` | Worker | 4 | 8192 | 125 | GPU Passthrough (`0000:28:00.0`) |

## Provisioning Workflow

The provisioning process involves:
1.  **Image Download**: Downloads the Ubuntu 24.04 Cloud Image (`ubuntu-main.img`) to local storage.
2.  **Cloud-Init Generation**: Generates a custom `cloud-init` configuration for each node, setting the hostname and user data.
3.  **VM Creation**: Creates Virtual Machines using the `l26` OS type, `virtio` network, and `scsi` storage.

### Provisioning Diagram

{% mermaid() %}
sequenceDiagram
    participant Pulumi
    participant ProxmoxAPI
    participant Storage
    participant VM

    Pulumi->>ProxmoxAPI: Authenticate
    Pulumi->>ProxmoxAPI: Download Ubuntu Cloud ISO
    ProxmoxAPI->>Storage: Save ubuntu-main.img
    
    loop For Each Node
        Pulumi->>Pulumi: Generate Cloud-Init Config
        Pulumi->>ProxmoxAPI: Upload Cloud-Init Snippet
        ProxmoxAPI->>Storage: Save snippet to local storage
        Pulumi->>ProxmoxAPI: Create VM
        ProxmoxAPI->>VM: Define Resources
        ProxmoxAPI->>VM: Attach Cloud-Init and Cloud Image
        ProxmoxAPI->>VM: Start VM
    end
{% end %}

## GPU Passthrough

Worker 4 is configured with PCI passthrough for a GPU.
```go
// Code snippet from proxmox.go
if config.HasGPU {
    // ...
    hostpcis = append(hostpcis, vm.VirtualMachineHostpciArgs{
        Device: pulumi.String("hostpci0"),
        Pcie:   pulumi.Bool(true),
        Id:     pulumi.String(pcieID), // 0000:28:00.0
        Xvga:   pulumi.Bool(i == 0),
    })
    // ...
}
```
