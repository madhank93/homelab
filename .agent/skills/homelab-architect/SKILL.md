---
name: homelab-architect
description: Architectural standards and workflows for the Homelab mono-repo.
---

# Homelab Architecture Standards

## Goal
To maintain a clean, consistent, and automated mono-repo for infrastructure and platform engineering.

## 1. Directory Structure
- **`infra/pulumi`**: The source of truth for all Infrastructure as Code.
  - `proxmox.go`: VM/Cluster logic.
  - `hetzner_vps.go`: Cloud logic.
  - `main.go`: Entry point, service selection (`k` config).
- **`.agent/skills`**: Knowledge base for the AI Agent.
- **`platform/`**: Ansible, CDK8s, or other config management (if used).

## 2. The "Just" Workflow
We use `just` (Justfile) as the single entry point for all operations.
- **Why**: Abstraction. The user shouldn't remember long Pulumi flags or Ansible paths.
- **Standard Commands**:
  - `just deploy-proxmox`: The robust, auto-recovering deploy (Up -> Refresh -> Up).
  - `just pulumi <stack> <action>`: Generic wrapper.
- **Rule**: If a workflow is complex (more than 1 command), WRAP IT in the Justfile.

## 3. Configuration Management
- **Loader**: Global `k` (using `knadh/koanf`) in `config.go`.
- **Source**: Reads `.env` and defaults.
- **Pattern**:
  - `InitConfig()` loads everything at start.
  - `os.Setenv` is used to pass secrets (like `PROXMOX_PASSWORD`) to Providers.
  - **Do NOT** hardcode secrets in Go files.

## 4. Verification Strategy
"Trust but Verify".
- **Infrastructure**: Verified by Pulumi (exit code 0).
- **Applications**: Verified by shell scripts (`verify_cluster.sh`).
  - Why? Pulumi `Helm` provider can sometimes stall if the cluster is not *fully* healthy yet. A script with `kubectl wait` loops is often more robust for initial bootstrap checks.

## 5. Agent Interaction
- **Artifacts**: Always keep `task.md` and `implementation_plan.md` up to date in the `brain/` directory.
- **Skills**: If a new pattern emerges (e.g., "Deploying MinIO"), create a new SKILL.md for it.
