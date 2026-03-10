package secrets

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewOpenBaoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Privileged namespace — OpenBao CSI provider DaemonSet requires elevated privileges
	// to mount secrets into pods via the CSI interface.
	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	// openbao-unseal-key — created by scripts/create-bootstrap-secrets.sh (not GitOps).
	// Populated from SOPS after first `bao operator init` run (see: just openbao-init).
	// Helm chart references this secret in extraContainers for auto-unseal on pod restart.

	cdk8s.NewHelm(chart, jsii.String("openbao-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("openbao"),
		Repo:        jsii.String("https://openbao.github.io/openbao-helm"),
		Version:     jsii.String("0.25.6"),
		ReleaseName: jsii.String("openbao"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// Disable sidecar injector — CSI only
			"injector": map[string]any{
				"enabled": false,
			},
			// Deploy OpenBao CSI provider DaemonSet (bridges OpenBao ↔ CSI driver)
			"csi": map[string]any{
				"enabled": true,
			},
			"server": map[string]any{
				"dataStorage": map[string]any{
					"enabled":      true,
					"size":         "10Gi",
					"storageClass": "longhorn",
				},
				"standalone": map[string]any{
					"enabled": true,
					"config": `ui = true
listener "tcp" {
  tls_disable     = 1
  address         = "[::]:8200"
  cluster_address = "[::]:8201"
}
storage "file" {
  path = "/openbao/data"
}`,
				},
				// Expose unseal key to the main container and unseal sidecar
				"extraSecretEnvironmentVars": []map[string]any{
					{
						"envName":    "OPENBAO_UNSEAL_KEY",
						"secretName": "openbao-unseal-key",
						"secretKey":  "unseal-key",
					},
				},
				// Unseal sidecar — runs alongside the OpenBao server container.
				// Uses extraContainers (not extraInitContainers) because the server
				// must be running before it can accept unseal requests.
				// On first boot: OpenBao is uninitialized — run `just openbao-init` manually.
				// On subsequent restarts: sidecar unseals automatically from the bootstrap secret.
				"extraContainers": []map[string]any{
					{
						"name":  "unseal",
						"image": "openbao/openbao:2.5.1",
						"command": []string{
							"sh", "-c",
							// Poll until OpenBao is reachable and sealed, then unseal.
							// Loop handles subsequent seals (e.g., leader election changes).
							`while true; do
  STATUS=$(bao status -format=json 2>/dev/null || echo '{"sealed":true}')
  if echo "$STATUS" | grep -q '"sealed":true'; then
    bao operator unseal "$OPENBAO_UNSEAL_KEY" 2>/dev/null || true
  fi
  sleep 15
done`,
						},
						"env": []map[string]any{
							{
								"name": "OPENBAO_UNSEAL_KEY",
								"valueFrom": map[string]any{
									"secretKeyRef": map[string]any{
										"name": "openbao-unseal-key",
										"key":  "unseal-key",
									},
								},
							},
							{
								"name":  "BAO_ADDR",
								"value": "http://127.0.0.1:8200",
							},
						},
					},
				},
			},
		},
	})

	// HTTPRoute — openbao.madhan.app → openbao:8200 (UI)
	cdk8s.NewApiObject(chart, jsii.String("openbao-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("openbao"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"openbao.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "openbao", "port": 8200, "weight": 1},
				},
			},
		},
	}))

	return chart
}
