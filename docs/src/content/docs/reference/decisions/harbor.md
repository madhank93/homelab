---
title: Harbor for the registry
description: Why Harbor over a plain Docker registry for private images and pull-through caching.
---
**Decision:** Harbor, on the LAN only, behind the shared Gateway.

**Context:** The cluster needs somewhere to push its own images and a local cache for Docker Hub and GHCR pulls.

| Option | For | Against |
|---|---|---|
| **Harbor** | Projects and RBAC, pull-through proxy cache, built-in vulnerability scanning, web UI | Many components (core, portal, registry, jobservice, database, Redis); its own admin account |
| Docker Distribution (`registry`) | One small container | No RBAC, scanning or UI; caching one upstream per instance |

**Consequences:**
- Its chart cannot mount extra volumes, so a `secret-sync` pod is needed to get the admin password from OpenBao.
- Registry and jobservice PVCs are RWX to avoid Multi-Attach deadlocks on rollouts.
