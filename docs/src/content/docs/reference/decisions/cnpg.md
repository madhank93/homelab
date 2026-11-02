---
title: CloudNativePG for app databases
description: Why n8n's Postgres is a CloudNativePG Cluster rather than the Bitnami subchart bundled with its Helm chart.
---
**Decision:** Run application Postgres as CloudNativePG `Cluster` resources in each app's namespace; disable bundled database subcharts.

**Context:** n8n's chart bundles a Bitnami Postgres subchart. Its password would have to come from chart values or OpenBao, and its lifecycle is tied to n8n's chart.

| Option | For | Against |
|---|---|---|
| **CloudNativePG operator** | Database upgrades independently of the app; operator generates and owns the credentials (`<name>-app` Secret); scales to replicas and backups through one resource | One more operator to run |
| Bitnami subchart | Nothing extra to install | Password lives in values or a hand-made Secret; upgrades ride the app chart |

**Consequences:**
- The database password is the one app credential not in OpenBao.
- n8n's cluster is a single instance, so there is no failover yet.
- A major Postgres bump needs a new cluster or `pg_upgrade`.
