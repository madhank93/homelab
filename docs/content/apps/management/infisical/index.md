+++
title = "Infisical (Replaced)"
description = "Infisical was the previous secrets management platform — replaced by OpenBao + Secrets Store CSI Driver."
weight = 30
+++

## Status: Replaced by OpenBao

Infisical was the original runtime secrets management platform for this homelab. It has been replaced by [OpenBao](../../../apps/secrets/openbao/) + [Secrets Store CSI Driver](../../../apps/secrets/csi-driver/).

## Why Infisical Was Replaced

| Issue | Detail |
|-------|--------|
| Operator CRD schema bugs | The Infisical operator's `InfisicalSecret` CRD schema omits `projectSlug` from `secretsScope`, which breaks ArgoCD's ServerSideApply structured merge diff engine |
| SA token discovery broken | Kubernetes auth flow broken on clusters running k8s 1.24+ due to changes in SA token projection |
| ArgoCD SSA workaround required | Every `InfisicalSecret` resource needed `argocd.argoproj.io/sync-options: ServerSideApply=false` to avoid diff errors |
| Features behind cloud plan | Key operator features required a paid Infisical Cloud subscription |

## Current Architecture

Runtime secrets are now managed by:

- **OpenBao** (Vault-compatible fork, MPL-2.0) — secrets store at `http://openbao.madhan.app`
- **Secrets Store CSI Driver** v1.5.6 — mounts OpenBao secrets as files into pods

See the [Secrets Management](../../../apps/secrets/) section for current documentation.
