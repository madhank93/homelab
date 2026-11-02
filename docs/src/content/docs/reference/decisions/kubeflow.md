---
title: Kubeflow without Istio
description: Why Kubeflow runs without Istio, Knative, KServe and Dex, and what replaces them.
---
**Decision:** Install Kubeflow's notebooks, pipelines, Katib, Trainer and Spark Operator from the upstream manifests, without Istio, Knative, KServe or Dex.

**Context:** Cilium already provides the Gateway, network policy and L7 visibility, and identity is handled at the Bifrost edge. Upstream Kubeflow assumes Istio for routing and auth.

| Component | Decision | Reason |
|---|---|---|
| Istio | Excluded | Cilium covers the mesh features in use; a second data plane doubles the networking to debug |
| Knative | Excluded | No serverless workloads |
| KServe | Excluded | Ollama serves the models in use |
| Dex / oauth2-proxy | Excluded | Single user; the route injects `kubeflow-userid` |

**Consequences:**
- Stub Istio CRDs keep controllers that watch Istio kinds from crashing, and a kustomize patch strips every Istio resource and upstream NetworkPolicy.
- The [notebook gateway controller](/workloads/ai/notebook-gateway-controller/) replaces the per-notebook VirtualServices with HTTPRoutes.
- No authentication: anyone on the LAN is `user@example.com`. Multi-user access needs an OIDC proxy.
