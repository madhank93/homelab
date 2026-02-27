+++
title = "Falco"
description = "Runtime security via eBPF syscall monitoring on Talos Linux."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/compliance/falco.go` |
| Namespace | `falco` |
| Helm chart | `falco` v8.0.0 / Falco 0.43 (falcosecurity.github.io/charts) |
| HTTPRoute | None |
| UI | No |

## Purpose

Falco monitors kernel syscalls on every node using eBPF. It detects runtime anomalies:
- Shell spawned inside a container
- Unexpected network connections
- Privilege escalation attempts
- File access to sensitive paths

Alerts are output as JSON to stdout and collected by the OTel Agent for forwarding to VictoriaLogs.

## Talos-Specific Configuration

Talos Linux requires specific Falco settings:

| Setting | Value | Reason |
|---------|-------|--------|
| `driver.kind` | `modern_ebpf` | Talos does not permit kernel module loading — `kmod` and `legacy_ebpf` drivers are unavailable |
| `driver.sysfsMountPath` | `/sys/kernel` | Exposes BTF at `/sys/kernel/btf/vmlinux` inside the container, required for CO-RE eBPF on Talos kernel 6.18+ |

## Chart Version Notes

Chart v8.0.0 uses **snake_case** config keys:
```yaml
json_output: true
grpc_output: {}
```

Earlier chart versions (< 8.0.0) used camelCase (`jsonOutput`, `grpcOutput`). Use the correct case for the chart version in use.

## Alert Output

Falco writes JSON alerts to stdout:

```json
{
  "output": "Warning Spawned a shell inside a container (user=root ...)",
  "priority": "Warning",
  "rule": "Terminal shell in container",
  "time": "2026-02-26T10:00:00Z"
}
```

These are collected by the OTel Agent DaemonSet and forwarded to VictoriaLogs.
