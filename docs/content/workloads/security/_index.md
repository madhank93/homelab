+++
title = "Security"
description = "Runtime security and vulnerability scanning: Falco, Kyverno, and Trivy."
weight = 30
sort_by = "weight"
+++

Three layers, at three different moments:

| | When it acts | What it does |
|---|---|---|
| [Kyverno](@/workloads/security/kyverno/index.md) | Admission | Rejects or mutates resources before they are created |
| [Trivy](@/workloads/security/trivy/index.md) | After admission, continuously | Scans running workloads for known CVEs |
| [Falco](@/workloads/security/falco/index.md) | Runtime | Watches syscalls and alerts on suspicious behaviour |

Only Kyverno is preventive. Trivy reports on what is already running, and Falco
detects behaviour after it happens.
