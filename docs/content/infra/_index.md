+++
title = "Infrastructure Overview"
weight = 10
sort_by = "weight"
+++

# Infrastructure Overview

This section documents the infrastructure details of the Homelab, managed using [Pulumi](https://www.pulumi.com/) with Go.

The infrastructure is divided into two main components:
1.  **Hetzner Cloud**: Hosts the "Bifrost" gateway/VPS.
2.  **Proxmox VE**: Hosts the local Kubernetes cluster.

## Architecture

The following diagram illustrates the high-level architecture of the infrastructure managed by Pulumi.

{% mermaid() %}
graph TD
    User([User]) --> |Internet| HetznerVPC
    
    subgraph HetznerVPC [Hetzner Cloud]
        Bifrost[Bifrost VPS]
        Firewall[Firewall Rules]
        Firewall --> Bifrost
    end

    subgraph HomeLab [Home Lab - Proxmox VE]
        ControlPlane[K8s Control Plane]
        subgraph CP_Nodes [Nodes]
            CP1[k8s-controller1]
            CP2[k8s-controller2]
            CP3[k8s-controller3]
        end
        ControlPlane --- CP_Nodes
        
        Workers[K8s Workers]
        subgraph Worker_Nodes [Nodes]
            W1[k8s-worker1]
            W2[k8s-worker2]
            W3[k8s-worker3]
            W4["k8s-worker4 (GPU)"]
        end
        Workers --- Worker_Nodes
    end

    Pulumi[Pulumi Engine] --> |API| HetznerVPC
    Pulumi --> |API| HomeLab
{% end %}

## Services

The infrastructure provisioning is controlled via a Go application that orchestrates deployments based on configuration.

### Deployment Flow

The `main.go` entrypoint loads configuration from `.env` and `config.yml` (using `koanf`) and selectively deploys enabled services.

- **Hetzner Service**: Deploys the Bifrost VPS and Firewall.
- **Proxmox Service**: Deploys the Kubernetes Cluster nodes using Cloud-Init.
