# Infisical Secrets Setup Guide

This document outlines the necessary steps to configure Infisical for the Homelab environment. These steps are critical for applications like Harbor, n8n, and Grafana to function correctly.

## Prerequisites

-   Access to the Infisical UI (e.g., `https://infisical.local` or via port-forward).
-   Infisical initialized with an admin account.

## Configuration Steps

### 1. Create Project
1.  Log in to Infisical.
2.  Create a new project named **`homelab-prod`**.

### 2. Verify Environment
1.  Ensure an environment named **`prod`** exists within the project.
    -   *Note: This is usually created by default.*

### 3. Create Secrets
Navigate to the **Secrets Dashboard** for the `prod` environment and create the following folders and secrets:

#### Path: `/harbor`
| Secret Key | Value | Description |
| :--- | :--- | :--- |
| `HARBOR_ADMIN_PASSWORD` | `<your-secure-password>` | Admin password for Harbor. |

#### Path: `/n8n`
| Secret Key | Value | Description |
| :--- | :--- | :--- |
| `DB_PASSWORD` | `<your-secure-password>` | Database password for n8n. |

#### Path: `/grafana`
| Secret Key | Value | Description |
| :--- | :--- | :--- |
| `ADMIN_PASSWORD` | `<your-secure-password>` | Admin password for Grafana. |

#### Path: `/cert-manager`
| Secret Key | Value | Description |
| :--- | :--- | :--- |
| `CLOUDFLARE_API_TOKEN` | `<your-cloudflare-api-token>` | API Token for DNS-01 challenges. |

### 4. Generate Service Token
1.  Go to **Project Settings** -> **Service Tokens**.
2.  Create a new token with the following settings:
    -   **Name**: `homelab-prod-token` (or similar).
    -   **Scopes**:
        -   `/harbor` (Read)
        -   `/n8n` (Read)
        -   `/grafana` (Read)
        -   `/cert-manager` (Read)
    -   **Expiration**: Set as desired (e.g., Never or 1 year).
3.  **Copy the generated token immediately.** You will not be able to see it again.

### 5. Apply Token to Kubernetes
Update the `infisical-service-token` Secret in the `infisical` namespace with the new token.

```bash
kubectl create secret generic infisical-service-token \
  --from-literal=infisicalToken="<PASTE_YOUR_TOKEN_HERE>" \
  -n infisical \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Troubleshooting
If applications fail to sync secrets:
1.  Verify the `InfisicalSecret` CRD status: `kubectl get infisicalsecret -A`
2.  Check the Infisical Operator logs: `kubectl logs -l app.kubernetes.io/name=infisical-operator -n infisical-operator-system`
