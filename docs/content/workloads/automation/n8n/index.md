+++
title = "n8n"
description = "Open-source workflow automation platform."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/automation/n8n.go` |
| Namespace | `n8n` |
| HTTPRoute | `n8n.madhan.app` → n8n service |
| UI | Yes |
| Requires Infisical | Yes — `n8n-db` Secret |

## Purpose

n8n is an open-source workflow automation platform. It integrates with 400+ services via nodes and supports scheduled, webhook, and event-driven workflows.

## Database

n8n uses PostgreSQL for workflow persistence. The database password is managed by Infisical at path `/n8n`, synced to the `n8n-db` Secret (`DB_PASSWORD` key).

## Encryption Key

n8n encrypts stored credentials with `N8N_ENCRYPTION_KEY`. This key must remain stable — changing it after workflows are created will break credential decryption.

> **Known issue:** If you see `Error: Credentials could not be decrypted`, the encryption key in the running Secret does not match the key used when credentials were first stored. See `docs/infisical-secrets-setup.md` for the recovery procedure.

## Accessing the UI

Navigate to `http://n8n.madhan.app`. On first access, create an admin account.
