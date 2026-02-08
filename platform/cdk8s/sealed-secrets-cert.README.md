# Sealed Secrets Public Certificate

This certificate is used by CI/CD to encrypt Kubernetes Secrets into SealedSecrets.

## How to Update This Certificate

If you rotate the Sealed Secrets controller keys, update this file:

```bash
# Fetch new cert
kubeseal --fetch-cert \
  --controller-name=sealed-secrets-controller \
  --controller-namespace=kube-system \
  > platform/cdk8s/sealed-secrets-cert.pem

# Commit
git add platform/cdk8s/sealed-secrets-cert.pem
git commit -m "Update sealed secrets public cert"
```

## Security Note

This is a **public certificate** - it's safe to commit to Git!
- ✅ Can encrypt secrets
- ❌ Cannot decrypt secrets (only the controller's private key can)
