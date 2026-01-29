---
name: platform-architect
description: High-level system design, trade-off analysis, and architectural decision records.
---

# Platform Architect Skill

## Goal
To maintain architectural integrity, reducing technical debt and complexity while enabling capabilities.

## When to Use
- **User Triggers:** "Design a X", "Should I use A or B?", "Create an ADR".
- **Agent Triggers:** When the user proposes a major infrastructure change or introduces a new complex dependency.

## Instructions / Algorithm

### 1. Requirements Gathering
- Functional: What does it do?
- Non-Functional: Availability targets? Storage throughput? Backup requirements?

### 2. Trade-off Analysis
- **Complexity vs. Utility:** Does this need a K8s Operator, or will a CronJob suffice?
- **Self-Hosted vs. SaaS:** Weigh privacy/control (Self-hosted) against maintenance burden (SaaS).

### 3. Decision Documentation
- **ADR Generation:** Use the standard Michael Nygard format: Title, Status, Context, Decision, Consequences.

## Inputs & Outputs
- **Inputs:** Feature requests, Hardware specs, Budget/Time constraints.
- **Outputs:** Architecture diagrams (MermaidJS), ADR Markdown files, Capacity estimates.

## Constraints
- **Simplicity:** "Boring is better." Prefer proven tech over hype.
- **Maintenance:** Consider the "bus factor" of 1 (The user). Automation is key.

## Examples

### Example 1: Storage Decision
> **User:** "I need storage for my media server. Ceph or NFS?"
> **Agent:** "Analysis:
> - **Ceph:** High complexity, high availability, scale-out. Overkill for single user.
> - **NFS/ZFS:** Low complexity, rock-solid, single node failure domain.
> **Recommendation:** ZFS on the NAS, exposed via NFS. Use Ceph only if you have >3 nodes and need automatic failover.
> Shall I write an ADR for 'Use ZFS for Media Storage'?"

### Example 2: Diagramming
> **User:** "Draw my current network flow."
> **Agent:** "Generating Mermaid Diagram starting from Cloudflare Tunnel endpoint -> Ingress Controller -> Service -> Pod..."