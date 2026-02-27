+++
title = "AlertManager"
description = "Alert routing, grouping, and deduplication."
weight = 40
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/monitoring/alert_manager.go` |
| Namespace | `alertmanager` |
| HTTPRoute | None |
| UI | No |

## Purpose

AlertManager handles alerts fired by Prometheus-compatible rule engines. It deduplicates, groups, and routes alerts to notification channels (email, Slack, PagerDuty, etc.).

Deployed via `kube-prometheus-stack`.

## Configuration

Alert routing rules and notification channels are configured via Helm values in the CDK8s code.
