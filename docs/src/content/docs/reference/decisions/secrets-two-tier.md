---
title: Two-tier secrets
description: Why bootstrap secrets live in SOPS and age while runtime secrets live in OpenBao behind the Secrets Store CSI driver.
---
**Decision:** SOPS with age for the few secrets needed before the cluster can serve secrets; OpenBao with the Secrets Store CSI driver for everything else.

**Context:** The repository and its manifests branch are public, so no plaintext secret may reach git. OpenBao runs in the cluster, so it cannot hold the secrets that build the cluster or unseal OpenBao itself.

| Option | For | Against |
|---|---|---|
| Everything in SOPS | No in-cluster service to run | Every app secret becomes a k8s Secret in rendered manifests; rotation needs a commit |
| Everything in OpenBao | One store | Chicken-and-egg: Pulumi and the unseal step need secrets before OpenBao exists |
| SOPS for bootstrap + OpenBao at runtime (chosen) | Bootstrap set stays small and reproducible from git; app secrets never appear in manifests and can be mounted as files | Two systems; OpenBao must be unsealed after every restart |

**Consequences:**
- Losing the age private key means losing the bootstrap secrets; back it up outside the repository.
- Each app needs a SecretProviderClass and an OpenBao role, and a pod must mount the CSI volume for a synced Secret to exist.
- The unseal key sits in a k8s Secret so a sidecar can unseal OpenBao unattended.
