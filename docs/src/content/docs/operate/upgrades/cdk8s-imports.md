---
title: CDK8s imports
description: Regenerate the typed Go bindings after changing any chart or the Kubernetes API version in cdk8s.yaml.
sidebar:
  order: 18
---

Charts with a generated typed package are pinned in [`workloads/cdk8s.yaml`](src:workloads/cdk8s.yaml). After changing any of them, regenerate the bindings so the Go code compiles against the new values schema.

```bash frame="terminal"
cd workloads
cdk8s import        # regenerates workloads/imports/*
go mod tidy
just synth          # must compile before you push
```

When Kubernetes moves a minor, update the `k8s@X.Y.Z` import at the top of `cdk8s.yaml` too: the API types must stay on the cluster's minor. Charts rendered through `cdk8s.NewHelm` have no typed package; they are pinned in `config/versions.yaml` and need no import.
