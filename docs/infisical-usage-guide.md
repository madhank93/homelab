# Managing Secrets with Infisical in Homelab

## Overview
Your cluster is now equipped with a centralized secret management system.
-   **Infisical Server**: Stores the secrets encrypted. Accessible at `https://infisical.local` (or via IP).
-   **Infisical Operator**: Runs in the cluster, fetches secrets from the server, and creates Kubernetes `Secret` objects.

## Workflow for New Applications

### 1. Store Secrets in Infisical
1.  Log in to the Infisical UI (`https://infisical.local`).
2.  Create a new **Project** for your application (e.g., `my-new-app`).
3.  Add your secrets (e.g., `API_KEY`, `DB_PASSWORD`) to the project.

### 2. Generate a Service Token
1.  In Infisical Project Settings, go to **Service Tokens**.
2.  Create a new token (e.g., "production-token").
3.  **Copy the token immediately**. You will not see it again.

### 3. Provide Access to the Operator
The operator needs permission to access *this specific project*.
Create a Kubernetes Secret containing the token in your app's namespace.

**Example CDK8s Code (`my-app.go`):**
```go
// 1. Create a Secret for the Service Token (Managed by SealedSecrets or Manually)
authSecret := cdk8s.NewApiObject(chart, jsii.String("auth-secret"), &cdk8s.ApiObjectProps{
    ApiVersion: jsii.String("v1"),
    Kind:       jsii.String("Secret"),
    Metadata: &cdk8s.ApiObjectMetadata{
        Name: jsii.String("infisical-auth-token"),
        Namespace: jsii.String("my-app-namespace"),
    },
    StringData: map[string]string{
        "token": "st.your-service-token-here...", // Ideally, seal this!
    },
})

// 2. Define the InfisicalSecret CRD
cdk8s.NewApiObject(chart, jsii.String("infisical-secret"), &cdk8s.ApiObjectProps{
    ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
    Kind:       jsii.String("InfisicalSecret"),
    Metadata: &cdk8s.ApiObjectMetadata{
        Name: jsii.String("my-app-secrets"),
        Namespace: jsii.String("my-app-namespace"),
    },
    Spec: map[string]interface{}{
        "hostAPI": "http://infisical-infisical-standalone-infisical.infisical.svc.cluster.local:8080", // Internal URL
        "authentication": map[string]interface{}{
            "serviceToken": map[string]interface{}{
                "serviceTokenSecretReference": map[string]interface{}{
                    "secretName": "infisical-auth-token",
                    "secretKey":  "token",
                },
            },
        },
        "managedSecretReference": map[string]interface{}{
            "secretName": "my-app-k8s-secret", // The native K8s secret name to create
        },
    },
})
```

### 4. Consume Secrets in Your App
Your deployment can now reference the native Kubernetes Secret created by the operator:

```go
// In your Deployment container spec:
EnvFrom: []corev1.EnvFromSource{
    {
        SecretRef: &corev1.SecretEnvSource{
            Name: jsii.String("my-app-k8s-secret"), // Matches managedSecretReference above
        },
    },
},
```

## Setup Verification
To verify the operator is working:
`kubectl get infisicalkeys -n <namespace>`
It should show `Synced` status.
