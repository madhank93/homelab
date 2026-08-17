+++
title = "Falco"
description = "Runtime security via eBPF syscall monitoring on Talos Linux."
weight = 10
+++

## What is Falco?

[Falco](https://falco.org/) is a cloud-native runtime security tool that monitors kernel syscalls on every node using eBPF. It detects anomalous behavior in real time — shell spawned inside a container, unexpected network connections, privilege escalation attempts, or access to sensitive file paths — and generates structured alerts.

## Why Falco?

Falco is the CNCF Graduated standard for Kubernetes runtime security. Unlike image scanning (Trivy), which detects vulnerabilities in static images, Falco detects active threats at runtime — when a container is actually doing something suspicious. The two tools are complementary.

| Tool | What it covers |
|------|---------------|
| Trivy | Known CVEs in container images (static) |
| Falco | Suspicious behavior at runtime (dynamic) |

## How It's Used Here

Falco runs as a DaemonSet on every node (including control plane, via `tolerations: [{operator: Exists}]`). Alerts flow through falcosidekick and are collected by the OTel Agent for forwarding to VictoriaLogs and Grafana.

**Alert pipeline:**

```
Falco DaemonSet (every node)
  → JSON alert to stdout
  → falcosidekick (collects from gRPC)
  → OTel Agent filelog receiver (from falcosidekick stdout logs)
  → VictoriaLogs
  → Grafana (LogQL query)
```

Source: {{ src(path="workloads/security/falco.go") }}

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `falco` | Privileged PSA required |
| Helm chart | `falco` | falcosecurity.github.io/charts |
| `driver.kind` | `modern_ebpf` | Required on Talos — see below |
| `driver.sysfsMountPath` | `/sys/kernel` | Exposes BTF at `/sys/kernel/btf/vmlinux` |
| `json_output` | `true` | Machine-readable alert format |
| `grpc.enabled` | `true` | falcosidekick connection |
| `falcosidekick.enabled` | `true` | Collects alerts from gRPC, exposes metrics |
| `falcosidekick.webui.enabled` | `true` | Alert dashboard UI |
| Falco resources limit | `1000m` / `1024Mi` | eBPF program compilation is CPU-intensive |
| HTTPRoute | `falco.madhan.app` → falcosidekick-ui:2802 | Gateway API |

## Talos-Specific Configuration

Talos Linux locks down kernel module loading — the `kmod` and `legacy_ebpf` Falco drivers cannot be used.

`modern_ebpf` uses CO-RE (Compile Once, Run Everywhere) eBPF programs that load without a pre-compiled kernel module. These programs use BTF (BPF Type Format) for kernel struct layout information.

```go
"driver": map[string]any{
    "kind": "modern_ebpf",
    // Mount /sys/kernel into container so probe can find BTF at /sys/kernel/btf/vmlinux.
    // Required on Talos — container's default sysfs does not expose /sys/kernel/btf.
    "sysfsMountPath": "/sys/kernel",
},
```

## What Falco Detects (Default Rules)

- Shell spawned inside a container (`bash`, `sh`, `zsh`)
- Package management tools running in containers (`apt`, `yum`, `apk`)
- Privilege escalation (`setuid`, capability changes)
- Unexpected outbound connections from sensitive containers
- Read access to sensitive files (`/etc/passwd`, `/etc/shadow`, SSL certificates)
- K8s audit events (if K8s audit log is forwarded to Falco)

## Chart Version Notes

The chart uses **snake_case** config keys:

```yaml
json_output: true
grpc_output:
  enabled: true
```

Chart 8 renamed these from camelCase (`jsonOutput`, `grpcOutput`); on an upgrade
from an older chart they must be rewritten or Falco silently ignores them.

## How It Connects

```
Kernel syscalls on every node
  → Falco modern_ebpf probe (CO-RE, uses BTF from /sys/kernel/btf/vmlinux)
  → Falco DaemonSet (evaluates rules)
  → JSON alert to stdout + gRPC to falcosidekick
  → falcosidekick (aggregates, exposes metrics at :2801/metrics)
  → OTel Agent filelog receiver (collects stdout logs)
  → VictoriaLogs (log storage)
  → Grafana (alert queries + dashboard)
  → VMAgent → VictoriaMetrics (sidekick metrics via ServiceMonitor)
```

## Falcosidekick UI retention

The UI stores events in Redis on a fixed 1 Gi volume, and two chart values keep
that from wedging:

| Value | Setting | Why |
|---|---|---|
| `webui.replicaCount` | `1` | The chart key is `replicaCount`; `replicas` is silently ignored and leaves the default of 2 |
| `webui.ttl` | `7d` | Without a TTL, events accumulate forever |

The TTL is not just about disk. Redis writes a full temporary copy before
renaming it over `dump.rdb`, so once the dataset passes half the volume no save
can complete: it fails with `No space left on device`, leaves the partial file
behind, and answers `MISCONF` instead of `PONG` — which blocks the UI's
`wait-redis` init container permanently.

## Troubleshooting

### eBPF Probe Not Loading on Talos

**Symptoms:** Falco pod logs show `failed to load BPF probe` or `BTF not found`.

**Diagnosis:**

```bash
kubectl logs -n falco -l app.kubernetes.io/name=falco --tail=50 | grep -i "btf\|ebpf\|probe"
```

**Fix:** Ensure `sysfsMountPath: /sys/kernel` is set and the container can access `/sys/kernel/btf/vmlinux`. On Talos, the default container sysfs does not expose kernel BTF — the hostPath mount through `sysfsMountPath` is required.

### Falco Pod Consuming Too Much CPU

**Symptoms:** Falco pod uses high CPU on a busy node.

**Why:** The eBPF probe evaluates every syscall against Falco's ruleset. On nodes with many containers or high syscall rates, this is CPU-intensive.

**Mitigation:** Tune rules to reduce the number of active syscalls being evaluated, or exclude noisy containers with `ec2_tag_check` or `container.name` filters.

### Alerting in Grafana

```logql
# Query Falco alerts in Grafana:
{namespace="falco"} | json | priority="WARNING"
{namespace="falco"} |= "Terminal shell in container"
```
