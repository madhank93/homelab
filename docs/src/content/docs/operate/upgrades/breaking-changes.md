---
title: Breaking changes
description: Known breaking changes per component and version boundary, and what to check for each.
sidebar:
  order: 19
---

Check this table before any bump. Each row is a change that broke, or would have broken, this cluster.

| Upgrade | Breaking change |
|---|---|
| Talos minor | Default Kubernetes moves with it; Kubernetes is pinned separately and `k8s@` in `cdk8s.yaml` must match its minor |
| Talos 1.13 → 1.14 | Defaults to Kubernetes 1.37, which Cilium does not support; leave workload isolation off (breaks Longhorn iSCSI) |
| Cilium any | `CiliumLoadBalancerIPPool` / `CiliumL2AnnouncementPolicy` field renames — check `cilium.go` |
| Cilium 1.18 → 1.19 | Datapath drops LB-IPAM traffic on unused ports; an LB IP equal to a node IP cuts that node off. Use 1.19.5+ and pass gate 4 |
| Cilium 1.19 → 1.20 | Needs Gateway API CRDs v1.6.1+ applied first; keep the experimental channel while any TLSRoute `v1alpha2` reader remains |
| Grafana 10 → 11 | Angular plugins removed; datasource API changed |
| Longhorn minor | One minor at a time only |
| OpenBao 2.7 | Removes the `file` storage backend; image held on 2.6.x |
| NetBird any | Server, reverse-proxy, agent and in-cluster peer on the identical version |
| Authentik any | Migration race: `bootstrap.sh` migrates first; if the health check times out, SSH in and force-migrate ([Bifrost](/operate/upgrades/bifrost/)) |
| Authentik 2026.5 → 2026.8 | Forwarded headers honoured only from `AUTHENTIK_LISTEN__TRUSTED_PROXY_CIDRS` (default `172.16.0.0/12` covers Traefik on `bifrost_net`). Refuses to skip a minor |
| netbird-agent restart | Check `ip rule show \| grep 7120`; missing rules = restart the agent again |
| Trivy any | `charts.trivy-operator` in `versions.yaml` drives the CRD bundle; the chart comes from `cdk8s.yaml` |
| ComfyUI any | Needs a CUDA 12.8+ tag for sm_120; the `cu128` line is discontinued |
| n8n 1.x → 2.x | One-way DB migration — dump the CNPG database first; Code node in task runners, Execute Command disabled, in-memory binary data removed |
| Kubeflow 1.11 → 26.03 | Repo renamed to `kubeflow/community-distribution`; dashboard under `applications/dashboard/`, notebooks under `applications/notebooks-v1/`; Service `centraldashboard` → `dashboard`; two Deployments need deleting for new selectors |
