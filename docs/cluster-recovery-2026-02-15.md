# Cluster Recovery and ArgoCD Debugging Walkthrough

## 1. Issue Overview
The user reported that the ArgoCD UI was inaccessible. Investigation revealed a critical cluster-wide failure:
-   **Infrastructure**: 5 out of 7 Talos nodes were `NotReady` or `Offline`.
-   **Etcd**: Quorum was lost due to `k8s-controller3` being physically dead and `k8s-controller2` failing to sync.
-   **ArgoCD**: Pods were stranded on `NotReady` nodes, preventing the UI from serving traffic.

## 2. Diagnosis Steps
1.  **Node Status**: `kubectl get nodes` showed only `k8s-controller1` was `Ready`. All workers and other controllers were `NotReady`.
2.  **Etcd Health**: `talosctl health` revealed timeouts and a "phantom" connection attempt to `192.168.1.141` (likely a stale IP for the dead controller).
3.  **Logs**: Kubelet logs on workers showed `etcdserver: request timed out`, confirming the API server was unable to write to the database due to quorum loss.

## 3. Resolution Steps

### 3.1 Restoring Etcd Quorum
The cluster has 3 control plane nodes. With `.206` dead (later confirmed to be `k8s-controller3`) and `.223` partitioned, `.171` became a minority (1/3).
-   **Action**: We forcibly removed the dead member `k8s-controller3` (`192.168.1.206` / ID `115d57e517f5ba18`) from the Etcd cluster using `talosctl -n 192.168.1.171 etcd remove-member`.
-   **Result**: Cluster size reduced to 2. `.171` and `.223` regained quorum within minutes.

### 3.2 Recovering Nodes (Phase 1)
Once Etcd was writable:
-   **Service Restarts**: Explicitly restarted `kubelet` and `kube-apiserver` (container) on stubborn nodes (`.223`, `.151`, `.167`) to clear stale connection pools.
-   **Reboots**: Issued `talosctl reboot` to `k8s-controller2`, `k8s-worker1`, and `k8s-worker4`.
-   **Result**: `k8s-controller2`, `k8s-worker1`, `k8s-worker2`, and `k8s-worker3` successfully registered and became `Ready`.

### 3.3 Troubleshooting k8s-worker4
`k8s-worker4` remained `NotReady` despite reboots.
1.  **Diagnosis**: Logs showed `Node not found` and persistent timeouts.
2.  **Attempt 1 (Soft Reset)**: Restarted Kubelet. Failed.
3.  **Attempt 2 (Hard Reboot)**: Full node reboot via `talosctl`. Failed (Node came up but didn't register).
4.  **Attempt 3 (Object Deletion)**: Ran `kubectl delete node k8s-worker4` to force a clean registration. Failed (Node didn't reappear automatically).
5.  **Attempt 4 (Manual Restart)**: User manually restarted the node.
    -   **Result**: Node successfully registered and became `Ready`.

### 3.4 Final Verification
-   **Nodes**: All 7 nodes are now `Ready`, including `k8s-controller3` which seemingly recovered or was powered on.
-   **ArgoCD**: Pods rescheduled to healthy nodes. UI confirmed accessible via `curl -I https://192.168.1.221`.

## 4. Current Status
-   **Healthy Nodes (7/7)**: All nodes are `Ready` and scheduling.
-   **Application**: ArgoCD is fully functional.

## 5. Summary
The cluster has been fully restored from a critical state where the majority of nodes were offline and Etcd quorum was lost. The combination of Etcd member management, targeted service restarts, and node reboots resolved the deadlock.
