---
title: GitOps with Pulumi and CDK8s
description: Why infrastructure runs from a laptop with Pulumi while workloads go through CDK8s, CI and Argo CD from a release manifests branch.
---
**Decision:** Pulumi (Go) for VMs, cluster, platform and edge, run only from a laptop. CDK8s (Go) for workloads, synthesized in CI to a `<release>-manifests` branch that Argo CD syncs.

**Context:** Infrastructure changes can destroy VMs, DNS or the Gateway and need bootstrap secrets; workload changes are frequent, reviewable and safe to reconcile continuously. The repository is public.

| Option | For | Against |
|---|---|---|
| Pulumi everywhere, from CI | One tool, one pipeline | CI would need every bootstrap secret; a bad merge could destroy infrastructure |
| Helm or Kustomize files synced directly | No synthesis step | Untyped YAML; shared values (domain, gateway, OpenBao address) repeated per app |
| Pulumi on a laptop + CDK8s through CI (chosen) | Typed Go on both sides sharing `config/config.yml`; CI holds no secrets; Argo CD shows and heals drift | Two toolchains; workloads see a change only after CI publishes |

**Consequences:**
- A human runs every `just core <stack> up`; nothing in CI can change infrastructure.
- The cluster runs whatever `manifests_branch` in `config/config.yml` names, so a release is rolled out by changing one line and running `just core platform up`.
- Synthesized manifests are public, so CI blocks any new rendered Secret payload.
