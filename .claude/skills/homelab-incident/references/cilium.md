# Cilium incidents

Config and the full list of Talos-specific settings: [Cilium](../../../../docs/src/content/docs/platform/cilium.md). Upgrade-time behaviour: [Cilium upgrades](../../../../docs/src/content/docs/operate/upgrades/cilium.mdx).

## Which Cilium component

| Symptom | Component | Datapath risk |
|---|---|---|
| Node `NotReady`, agent `0/1`, CrashLoop | Agent (DaemonSet `cilium`) | Yes |
| LB IP dark on the LAN only | Agent on the L2 lease holder | Yes |
| New route status-less, 404 | Operator (`cilium-operator`) | No |

## Stale L2 lease

Fingerprint: from the LAN TCP to `192.168.1.220` times out; from a pod the Gateway answers. The holder keeps renewing and `Programmed=True` stays set, so nothing reports unhealthy. Ignore `Proxy Status: 0 redirects active`: Envoy runs as the external `cilium-envoy` DaemonSet and is never counted there.

```bash
kubectl get leases -n kube-system | grep l2announce
kubectl -n kube-system exec ds/cilium -- cilium-dbg shell -- db/show l2-announce
```

- Gateway: `just gateway-repair-l2`.
- Any other LoadBalancer Service (Ollama, CS2): `kubectl delete lease -n kube-system cilium-l2announce-<namespace>-<service>`.
- A ~35 s gap on every agent rollout while the holder's agent restarts is expected and self-heals; act only if it persists.

## Agent CrashLoop or wedged datapath

1. Confirm which node: `kubectl get pods -n kube-system -l app.kubernetes.io/name=cilium-agent -o wide`.
2. Deleting the pod does not clear corrupted BPF state. Ask the user, then reboot: `talosctl --talosconfig ~/.talos/config reboot --nodes <ip>`.
3. Reboot hangs at `unmountPodMounts` (an RWX share-manager ClusterIP is unreachable without a datapath): power-cycle per [Repairs](../../../../docs/src/content/docs/operate/repairs.md#node-hangs-on-reboot). Safe once kubelet and cri have stopped; Longhorn rebuilds from replicas.
4. After any agent restart, re-check the Gateway IP from the LAN.

If `just core platform up` reported the Cilium release failed because one agent never went ready, fix the node and have the user re-run it; it is idempotent.

## Stale operator Gateway watch

Fingerprint: after apiserver restarts (control-plane upgrade, controller reboot) a new HTTPRoute gets no `status.parents` and its host 404s, while existing routes work.

```bash
kubectl get gateway homelab-gateway -n kube-system -o jsonpath='{.status.listeners[*].attachedRoutes}'
kubectl get httproute -A --no-headers | wc -l
kubectl get lease cilium-operator-resource-lock -n kube-system -o jsonpath='{.spec.holderIdentity}'
kubectl get pods -n kube-system -l app.kubernetes.io/name=cilium-operator -o wide
```

The holder identity starts with the leader's node name; delete the operator pod on that node. The standby takes over and the route is Accepted within seconds. No datapath impact.

## Never

- Add `wt0` to `devices` in `core/platform/cilium.go`; see [Traffic](../../../../docs/src/content/docs/concepts/traffic.mdx).
