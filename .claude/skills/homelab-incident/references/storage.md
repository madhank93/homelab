# Storage incidents (Longhorn, CSI after pressure)

Component page: [Longhorn](../../../../docs/src/content/docs/workloads/storage/longhorn.md). Recipes: [Repairs](../../../../docs/src/content/docs/operate/repairs.md).

## Scope first

```bash
kubectl get engineimages -n longhorn-system           # STATE must be deployed
kubectl get nodes.longhorn.io -n longhorn-system      # every worker Ready, schedulable
kubectl get volumes.longhorn.io -n longhorn-system    # STATE / ROBUSTNESS
kubectl get pods -A --field-selector=status.phase=Failed
```

| Scope | Cause | Fix |
|---|---|---|
| Every attach cluster-wide, EngineImage `deploying`, "N of M" nodes | longhorn-manager scheduled on a node the engine-image DaemonSet cannot reach (a toleration without the matching `taintToleration`) | Code: keep `longhornManager.tolerations` in `workloads/storage/longhorn.go` matched by `defaultSettings.taintToleration`; never the control-plane taint |
| Every attach on one node, after a Talos upgrade, volumes cycle `faulted` | open-iscsi rejects an old record in `/var/lib/iscsi/nodes/` | `just longhorn-repair-iscsi` |
| One node, after DiskPressure; `ManagerPodDown`, CSI `driver ... not found` | DaemonSet pods stuck `Failed` | `kubectl delete pods -A --field-selector=status.phase=Failed`, then re-check Longhorn nodes |
| One volume | Wedged attachment or RWX share-manager | [Longhorn troubleshooting](../../../../docs/src/content/docs/workloads/storage/longhorn.md#troubleshooting) |
| `Multi-Attach error` during a rollout | Old pod still holds the RWO volume | [Repairs](../../../../docs/src/content/docs/operate/repairs.md#rwo-volume-deadlocked-during-a-rollout); confirm before force-deleting |

## Rules

- `defaultSettings.taintToleration` cannot change while volumes are attached; fix the manager toleration instead.
- `kubectl patch` on Longhorn DaemonSets or the Application is reverted by the `cots-applications` ApplicationSet. Change the code.
- Detaching a volume or force-deleting its pod needs user confirmation.
- `ROBUSTNESS=degraded` alone is not an incident: Longhorn rebuilds once the replica node is back.
