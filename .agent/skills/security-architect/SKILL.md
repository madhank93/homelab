---
name: security-architect
description: Expert security guidance: Zero Trust, OIDC (Authentik), Secrets (Infisical), and Scanning (Trivy).
---

# Security Architect Skill

## Goal
Enforce security posture using the installed toolset: Authentik (Identity), Infisical (Secrets), and Trivy (Scanning).

## When to Use
- **User Triggers:** "Secure this", "Rotate secrets", "Scan for vulns", "Configure SSO".
- **Agent Triggers:** When reviewing manifests or architecture.

## Instructions / Algorithm

### 1. Secrets Management (Infisical)
- **Pattern:** Do not hardcode secrets in CDK8s or Pulumi.
- **Usage:** Retrieve secrets from Infisical at runtime or sync via ExternalSecrets/Infisical Operator to K8s Secrets.
- **Code:** Ensure `seccomp.NewInfisicalChart` is running and managing the relevant stores.

### 2. Identity (Authentik)
- **Configuration:** Managed via Pulumi (`infra/pulumi/authentik_idp.go`).
- **Integration:** Prefer OIDC sidecars or native app integration pointing to Authentik issuer.

### 3. Vulnerability Scanning (Trivy)
- **Runtime:** Trivy is deployed (`cdk8s/cots/seccomp`).
- **Check:** Verify Trivy Operator reports for running pods.
- **CI/CD:** Suggest scanning images before deployment if building locally.

## Inputs & Outputs
- **Inputs:** Architecture, Secret requirements.
- **Outputs:** Policy definitions, Secret injection patterns, Auth flows.

## Constraints
- **Zero Trust:** Assume network is hostile.
- **Least Privilege:** Scoped API keys for Infisical.

## Examples

### Example 1: Injecting a Secret
> **User:** "My app needs a database password."
> **Agent:** "Do not add it to `env` vars in plain text.
> 1.  Add the secret to Infisical dashboard.
> 2.  In CDK8s, configure the InfisicalSecret CRD (or ExternalSecret) to sync it to a K8s Secret.
> 3.  Mount the K8s Secret in your container."