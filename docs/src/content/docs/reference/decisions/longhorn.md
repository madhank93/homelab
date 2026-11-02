---
title: Longhorn for storage
description: Why replicated storage on the workers' OS disks beat Rook-Ceph, NFS and local-path.
---
**Decision:** Longhorn, 3 replicas across the 4 workers, as the default StorageClass.

**Context:** The workers are Proxmox VMs with one OS disk each and no spare block devices, and stateful apps must survive losing a node.

| Option | For | Against |
|---|---|---|
| **Longhorn** | Uses part of each OS disk; RWX through a built-in NFS share-manager; UI for volumes and snapshots | Single vendor; replicas share disk I/O with the OS |
| Rook-Ceph | Feature-rich, scales well | Needs dedicated disks per node; heavy for 4 small workers |
| NFS server | Simple | Single point of failure, no replication |
| local-path | No overhead | Data is tied to one node |

**Consequences:**
- Capacity is limited by OS disk free space, so scheduling relies on 300% overprovisioning.
- `longhorn-manager` must stay off control-plane nodes, or engine images never finish deploying and attaches block.
- Talos upgrades can leave iSCSI records that block attaches; `just longhorn-repair-iscsi` clears them.
