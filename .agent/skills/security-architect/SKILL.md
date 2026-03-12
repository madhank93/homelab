---
name: security-architect
description: Expert security guidance: Zero Trust, OIDC (Authentik), Secrets (OpenBao + CSI Driver), and Scanning (Trivy).
---

# Security Architect Skill

## Goal
Enforce security posture using the installed toolset: Authentik (Identity), OpenBao (Secrets), and Trivy (Scanning).

## When to Use
- **User Triggers:** "Secure this", "Rotate secrets", "Scan for vulns", "Configure SSO".
- **Agent Triggers:** When reviewing manifests or architecture.

## Instructions / Algorithm

### 1. Secrets Management (OpenBao + CSI Driver)
- **Pattern:** Do not hardcode secrets in CDK8s or Pulumi. CDK8s must generate zero `Secret` resources.
- **Usage:** Secrets are fetched from OpenBao at runtime via the Secrets Store CSI Driver.
- **Pattern A (file-only):** Mount secrets as files (`/mnt/secrets/<KEY>`). Use `GF_SECURITY_ADMIN_PASSWORD__FILE` style env vars. Used by Grafana.
- **Pattern B (secretObjects sync):** SecretProviderClass with `secretObjects` syncs a k8s Secret. Used by Harbor, n8n, Rancher, NetBird when the Helm chart requires `existingSecret`.
- **Code:** Ensure `secrets.NewOpenBaoChart` and `secrets.NewCsiDriverChart` are running. Add a `SecretProviderClass` per app namespace.

### 2. Identity (Authentik)
- **Configuration:** Managed via Pulumi (`core/cloud/authentik.go`).
- **Integration:** Prefer OIDC sidecars or native app integration pointing to Authentik issuer (`https://auth.madhan.app`).

### 3. Vulnerability Scanning (Trivy)
- **Runtime:** Trivy is deployed (`workloads/security/trivy.go`).
- **Check:** Verify Trivy Operator reports for running pods.
- **CI/CD:** Suggest scanning images before deployment if building locally.

## Inputs & Outputs
- **Inputs:** Architecture, Secret requirements.
- **Outputs:** Policy definitions, Secret injection patterns, Auth flows.

## Constraints
- **Zero Trust:** Assume network is hostile.
- **Least Privilege:** Scoped OpenBao K8s auth roles per app ServiceAccount.

## Examples

### Example 1: Injecting a Secret
> **User:** "My app needs a database password."
> **Agent:** "Do not add it to `env` vars in plain text.
> 1.  Add the secret to OpenBao: `bao kv put secret/myapp DB_PASSWORD=<value>`
> 2.  In CDK8s, add a `SecretProviderClass` (Pattern A or B depending on how the app reads secrets).
> 3.  Mount the CSI volume in your container. For Pattern B, the k8s Secret is auto-synced."
