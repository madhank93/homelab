---
name: sensitive-data-protection
description: Strict guidelines for protecting user privacy and preventing access to sensitive files (keys, secrets, credentials).
---

# Sensitive Data Protection

This skill defines strict rules for the agent to respecting user privacy and protecting sensitive data.

## 🚫 STRICTLY PROHIBITED ACTIONS (AT ANY COST)

The agent must **NEVER** attempt to read, cat, grep, access, or **retrieve** the contents of the following file types, paths, or **remote resources**, unless explicitly and forcefully instructed by the user in a follow-up confirmation:

1.  **SSH Keys**:
    *   `~/.ssh/id_*` (e.g., `id_rsa`, `id_ed25519`)
    *   `**/*.pem`, `**/*.key`, `**/*.ppk`

2.  **Kubernetes Secrets & Configs (Local & Remote)**:
    *   `~/.kube/config` (if it contains embedded certs/keys)
    *   `*.kubeconfig`
    *   **Retrieving Secrets**: `kubectl get secret ... -o jsonpath="{.data.password}"`, `kubectl gets secret ... -o yaml` (if it dumps data).
    *   **Decoding Secrets**: PIPING any secret output to `base64 -d` is **STRICTLY FORBIDDEN**.

3.  **Environment & Config Secrets**:
    *   `.env` files usually contain API keys. Do not read them.
    *   `passwd`, `shadow` files.
    *   `docker-compose.yml` (if it contains plaintext environment variables for passwords/secrets - check first or ask).

4.  **Pulumi & Infrastructure State**:
    *   `Pulumi.*.yaml` (stack configs) often contain encrypted secrets (ciphertext). While technically safe to read ciphertext, avoid it to prevent accidental diffs or leakage.
    *   Local state files (e.g., in `.pulumi`).

5.  **Encryption Keys**:
    *   Any file resembling a private key or certificate (`.p12`, `.pfx`, `.crt` + `.key`).

## ✅ SAFE PRACTICES

*   **Assumption of Presence**: Assume necessary keys (SSH, Cloud creds) are **already loaded** in the environment (e.g., `ssh-agent`, env vars). Do not try to verify them by reading the files.
*   **Verification**: If you need to verify a key exists, use `ls` or `stat`, but **NEVER** `cat` or `read`.
*   **User Prompt**: If a tool fails due to missing credentials, **ask the user** to permit the action or provide the credential safely, rather than trying to find it yourself.
*   **Logs**: Be careful when reading logs that might dump secrets.

## 🛡️ ENFORCEMENT

*   **CLI Command Review**: Before running ANY command (run_command), check if it outputs a secret. If it does (e.g., `cat .env`, `kubectl get secret -o yaml`), **ABORT**.
*   **Artifact Review**: Do not write secrets into artifacts (walkthroughs, plans).
*   If a user request implies reading these files (e.g., "Check my SSH config"), **ask for permission** first and clarify you will only look at non-sensitive parts (like `config` vs `id_rsa`).
