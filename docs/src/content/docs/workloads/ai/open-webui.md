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

Open WebUI has its own login, not Authentik. Sign-up is off (`ENABLE_SIGNUP=false`), so the admin creates accounts under **Admin Panel → Users**. On a fresh PVC with no users, the sign-up form still appears once so the first account can be created and become admin.

The SQLite database, uploads and the session secret key it generates on first start all live on the PVC, so a pod restart keeps every account and chat.

## Settings come from env, not the admin UI

`ENABLE_PERSISTENT_CONFIG=false` makes the env vars in `open_webui.go` the only source of settings. Changes made under **Admin Panel → Settings** work until the pod restarts, then reset. To keep a setting, add its env var to the Deployment ([env reference](https://docs.openwebui.com/getting-started/env-configuration)).

Without this flag, settings saved in the database override the env vars, and changing an env var does nothing.

## Built-in tools are off

Open WebUI 0.12 attaches its built-in tools (tasks, memories, skills, knowledge) to every chat by default. The local Qwen3-Coder Q3 models answer with bare tool-call JSON that Ollama does not parse as a tool call, so a plain "hi" came back as `{"name": "create_tasks", ...}`. `DEFAULT_MODEL_METADATA={"capabilities":{"builtin_tools":false}}` turns them off for every model.

## Picking a model

- **`executor`** has a 28K context, but its system prompt is tuned for aider (code only, no prose), so it is terse in chat.
- **`hf.co/unsloth/Qwen3-Coder-…`** chats normally, but runs at Ollama's default 2K context, so long conversations get truncated.

The first message after a few idle minutes takes about 45s while Ollama loads the model back onto the GPU. Switching between these two models reloads it too: they share weights but load separately.

## Keeping it private

`chat` is not in `publicServices` ([Exposure](/cloud/exposure/)), so it resolves only to the LAN gateway `192.168.1.220`. Do not add it there without putting Authentik in front first.
