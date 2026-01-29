---
name: sre-operator
description: Observability with VictoriaMetrics/components, incident response, and reliability.
---

# SRE Operator Skill

## Goal
To maintain reliability using the VictoriaMetrics ecosystem (VM, VictoriaLogs) and Grafana.

## When to Use
- **User Triggers:** "Create a dashboard", "Query logs", "Debug latency", "Alert on errors".
- **Agent Triggers:** When debugging workloads or defining monitoring rules.

## Instructions / Algorithm

### 1. Metrics (VictoriaMetrics)
- **Querying:** Use MetricsQL (PromQL compatible).
- **Storage:** Data is stored in the `victoria-metrics` app (deployed via CDK8s).
- **Agents:** Ensure `vmagent` or Prometheus scrape configs are present.

### 2. Logs (VictoriaLogs)
- **Querying:** Use LogsQL.
- **Pattern:** `_stream:{label="value"}` to filter streams.
- **Interaction:** Suggest checking Grafana Explore view for logs.

### 3. Alerting (AlertManager)
- Define `VMAlert` rules or standard Prometheus rules.
- Route criticals to PagerDuty/Email via AlertManager config (`cdk8s/cots/monitoring`).

## Inputs & Outputs
- **Inputs:** MetricsQL/LogsQL queries, Error rates.
- **Outputs:** Dashboard JSON, Rule definitions.

## Constraints
- **Performance:** Watch cardinality on metrics.
- **Retention:** Be playful but mindful of disk on `victoria-metrics` PVCs.

## Examples

### Example 1: Log Query
> **User:** "Show me logs for the 'harbor' app."
> **Agent:** "Using LogsQL for VictoriaLogs:
> `_stream:{app="harbor"} | limit 100`
> This filters the stream for the app label 'harbor'."