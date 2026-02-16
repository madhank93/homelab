# Infisical Implementation Report

## 1. Overview
This document details the successful implementation of a self-hosted Infisical instance on Kubernetes using `cdk8s`. The solution deploys Infisical as a standalone service backed by a dedicated PostgreSQL database and Redis, exposed via the Cilium Gateway API.

## 2. Architecture
- **Application**: Infisical Standalone v1.7.2
- **Database**: Bitnami PostgreSQL 16.2.5 (deployed separately)
- **Cache**: Embedded Redis (within Infisical chart)
- **Ingress**: Gateway API (`HTTPRoute`) via Cilium
- **Storage**: Longhorn (10Gi for Postgres)
- **Secrets**: Kubernetes Secrets (managed via GitOps/SealedSecrets pattern)

## 3. Implementation Details

### 3.1 Secrets Management
To ensure security and proper GitOps practices, sensitive data is injected via environment variables at runtime and stored in a Kubernetes Secret.

- **Source**: `INFISICAL_DB_PASSWORD`, `INFISICAL_ENCRYPTION_KEY`, `INFISICAL_AUTH_SECRET` (read from env).
- **Resource**: `Secret/infisical-secrets`
- **Keys**:
  - `DB_PASSWORD`: Used by both Postgres (init) and Infisical (connect).
  - `ENCRYPTION_KEY`: 16-byte hex string for Infisical platform encryption.
  - `AUTH_SECRET`: Used for signing tokens.
  - `DB_CONNECTION_URI`: Pre-calculated connection string injected to prevent default/root user assumptions.

#### Generating Secrets
Run the following commands to generate secure values for your `.env` or CI/CD secrets:

```bash
# INFISICAL_DB_PASSWORD (random string)
openssl rand -hex 16

# INFISICAL_ENCRYPTION_KEY (MUST be exactly 16 bytes / 32 hex chars)
openssl rand -hex 16

# INFISICAL_AUTH_SECRET (random signing key)
openssl rand -base64 32
```

```go
	// Create full connection URI explicitly to avoid chart defaults
	dbConnectionUri := "postgresql://infisical:" + infisicalDbPassword + "@postgresql:5432/infisical"
```

### 3.2 PostgreSQL Deployment
We decoupled PostgreSQL from the Infisical chart to gain fine-grained control over the deployment, specifically for image management and persistence.

- **Chart**: `bitnami/postgresql` (Version `16.2.5`)
- **Image Tag**: Pinned to `latest` to resolve `ImagePullBackOff` errors encountered with specific revision tags (`17.2.0-debian-12-r2`, `16`).
- **Persistence**: Enabled with `longhorn` storage class (10Gi).
- **Network Policy**: Explicitly disabled (`enabled: false`) to prevent connectivity issues during initialization.

```go
		"image": map[string]any{
			"tag": "latest",
		},
```

### 3.3 Infisical Configuration
The Infisical chart was configured to connect to the external Postgres instance.

- **Database**: Configured to use `existingConnectionStringSecret` pointing to `infisical-secrets`.
- **Resources**:
  - Requests: 200m CPU, 512Mi Memory
  - Limits: 1024Mi Memory
- **Replica Count**: 1

### 3.4 Ingress & Routing
Access is managed via the Gateway API.

- **Resource**: `HTTPRoute/infisical`
- **Gateway**: `homelab-gateway` (Namespace: `kube-system`)
- **Hostnames**:
  - `infisical.madhan.app`: Primary domain (HTTPS listeners on gateway may conflict with HTTP routing).
  - `infisical.local`: Alternative hostname added to verify HTTP access and bypass Gateway listener conflicts.
- **Port**: Routes to service port `8080`.

## 4. Verification

### 4.1 Pod Health
- **PostgreSQL**: `1/1 Running`. Logs confirm database initialization and readiness.
- **Infisical**: `1/1 Running`. Logs confirm successful migration (`Finished application migrations`) and server startup.

### 4.2 Accessibility
- **Internal**: `curl localhost:8080` (via port-forward) returns `200 OK`.
- **External (Gateway)**:
  - `curl -H "Host: infisical.local" http://192.168.1.220` returns `200 OK`.
  - `infisical.madhan.app` is currently routed but may require HTTPS configuration at the Gateway level for browser access without local host file overrides.

## 5. Usage
To access the application:
1. Add `192.168.1.220 infisical.local` to your `/etc/hosts`.
2. Navigate to `http://infisical.local`.
3. Create the initial admin account.
