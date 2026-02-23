# Sealed Secrets Key Rotation Runbook

## Background

The `sealed-secrets-controller` holds a **private key** in `kube-system` (as a
`kubernetes.io/tls` Secret labelled `sealedsecrets.bitnami.com/sealed-secrets-key`).
All SealedSecrets in the cluster are encrypted with the corresponding **public
certificate**, which is stored at `platform/cdk8s/sealed-secrets-cert.pem`.

> ⚠️ If the controller is ever re-created (e.g. due to ImagePullBackOff, cluster
> rebuild, or key rotation), it generates a **new** keypair. Existing SealedSecrets
> encrypted with the old cert will fail with `no key could decrypt secret` until
> they are re-sealed with the new cert.

---

## Symptoms

- ArgoCD shows apps as `Degraded`.
- Pods fail with `Error: secret "..." not found` or `CreateContainerConfigError`.
- `kubectl logs -n kube-system sealed-secrets-controller-...` shows:
  ```
  ErrUnsealFailed: no key could decrypt secret (...)
  ```

---

## Recovery Steps

### 1. Verify the controller is running

```bash
kubectl get pods -n kube-system -l app.kubernetes.io/name=sealed-secrets
```

If it is in `ImagePullBackOff` or `CrashLoopBackOff`, fix that first (see below).

### 2. Check what private keys exist

```bash
kubectl get secrets -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key
```

- **Multiple keys** → the old key still exists; the controller should auto-decrypt.
  Check if the controller is simply restarting or misconfigured.
- **Only one very recent key** → the old private key is gone; proceed to step 3.

### 3. Fetch the new public certificate

```bash
kubeseal --fetch-cert \
  --controller-name=sealed-secrets-controller \
  --controller-namespace=kube-system \
  > platform/cdk8s/sealed-secrets-cert.pem
```

> This only fetches the **public certificate** — safe to commit to Git.

### 4. Commit and push to trigger re-sealing

```bash
git add platform/cdk8s/sealed-secrets-cert.pem
git commit -m "fix: rotate sealed-secrets public cert after controller key loss"
git push
```

The `cdk8s-seal-publish.yml` GitHub Actions workflow will:
1. Synthesize manifests from `platform/cdk8s/`
2. Re-seal all secrets using the new cert
3. Publish to the `{branch}-manifests` branch
4. ArgoCD syncs automatically

---

## Fixing a Bad Image on the Controller

If the controller itself cannot start (e.g. `ImagePullBackOff`):

```bash
# Quick live patch (buys time while the proper fix is committed)
kubectl patch deployment sealed-secrets-controller -n kube-system \
  --type='json' \
  -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"ghcr.io/bitnami-labs/sealed-secrets-controller:0.35.0"}]'
```

The permanent fix is in `platform/cdk8s/cots/seccomp/sealedsecrets.go` — ensure
the `image` Helm value sets `registry: ""` to prevent Helm from prepending `docker.io/`
to the `ghcr.io` repository:

```go
"image": map[string]any{
    "registry":   "",   // prevents Helm prepending docker.io/
    "repository": "ghcr.io/bitnami-labs/sealed-secrets-controller",
    "tag":        "0.35.0",
},
```

---

## Incident: 2026-02-23

- **Cause**: Sealed-secrets controller was in `ImagePullBackOff` due to malformed
  image URI `docker.io/ghcr.io/...` (Helm default registry prepended to `ghcr.io` repo).
- **Impact**: All SealedSecrets across all namespaces failed to decrypt. Infisical,
  Harbor, n8n, and alertmanager apps were all `Degraded` in ArgoCD.
- **Fix**:
  1. Patched live deployment image to `ghcr.io/bitnami-labs/sealed-secrets-controller:0.35.0`.
  2. Updated `sealed-secrets-cert.pem` with new controller's public cert.
  3. Fixed `sealedsecrets.go` to set `registry: ""` — permanent fix via pipeline.
