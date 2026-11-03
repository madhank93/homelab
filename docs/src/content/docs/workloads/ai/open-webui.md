---
title: Open WebUI
description: Browser chat for the models served by Ollama. LAN only.
sidebar:
  order: 25
---

Open WebUI is a ChatGPT-style front end for the in-cluster [Ollama](/workloads/ai/ollama/). [Upstream docs](https://docs.openwebui.com/).

| | |
|---|---|
| Namespace | `open-webui` |
| Source | [`open_webui.go`](src:workloads/ai/open_webui.go) |
| Exposure | LAN · `chat.madhan.app` |
| Storage | 10Gi Longhorn PVC at `/app/backend/data` |

It needs no GPU, so it runs on the general workers and talks to Ollama at `ollama.ollama.svc.cluster.local:11434`. Models pulled into Ollama appear in its model picker.

## Accounts

Open WebUI has its own login, not Authentik. The first account to sign up becomes admin. Later sign-ups land as `pending` (`DEFAULT_USER_ROLE`) until the admin approves them under **Admin Panel → Users**.

The SQLite database, uploads and the session secret key it generates on first start all live on the PVC, so a pod restart keeps every account and chat.

## Keeping it private

`chat` is not in `publicServices` ([Exposure](/cloud/exposure/)), so it resolves only to the LAN gateway `192.168.1.220`. Do not add it there without putting Authentik in front first.
