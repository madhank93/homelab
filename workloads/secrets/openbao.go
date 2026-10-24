package secrets

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewOpenBaoChart deploys OpenBao (open-source Vault fork) as the cluster
// secrets manager. A sidecar auto-unseals it on restart from the bootstrap
// Secret openbao-unseal-key (created by scripts/create-bootstrap-secrets.sh).
func NewOpenBaoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// The OpenBao CSI provider DaemonSet needs host access.
	kit.Namespace(chart, namespace, true)

	openbao := config.HelmChart("openbao")
	cdk8s.NewHelm(chart, jsii.String("openbao-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("openbao"),
		Repo:        jsii.String(openbao.Repo),
		Version:     jsii.String(openbao.Version),
		ReleaseName: jsii.String("openbao"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// Secrets reach pods through the CSI provider only.
			"injector": map[string]any{
				"enabled": false,
			},
			"csi": map[string]any{
				"enabled": true,
				"agent": map[string]any{
					"image": map[string]any{"tag": config.ContainerImage("openbao").Tag},
				},
			},
			"server": map[string]any{
				// Pinned to 2.6.x, not the chart appVersion: 2.7 drops the
				// `file` storage backend configured below.
				"image": map[string]any{"tag": config.ContainerImage("openbao").Tag},
				"podAnnotations": map[string]any{
					"reloader.stakater.com/auto": "true",
				},
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
  telemetry {
    unauthenticated_metrics_access = true
  }
}
# In-pod only (kubectl exec): the one listener that allows the unauthenticated
# generate-root used by "just openbao-token"; disabled by default since 2.5.3.
listener "unix" {
  address                                  = "/tmp/openbao.sock"
  disable_unauthed_generate_root_endpoints = false
}
storage "file" {
  path = "/openbao/data"
}
# Declarative: API audit creation is off by default (unsafe_allow_api_audit_creation).
audit "file" "stdout" {
  options {
    file_path = "stdout"
  }
}
telemetry {
  prometheus_retention_time = "30s"
  disable_hostname          = true
}`,
				},
				// A sidecar, not an init container: the server must be running to
				// accept unseal requests. First boot still needs `just openbao-init`.
				// Only this container gets the unseal key.
				"extraContainers": []map[string]any{
					{
						"name":  "unseal",
						"image": config.ContainerImage("openbao").Ref(),
						"command": []string{
							"sh", "-c",
							// Loops so later seals are also undone.
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

	kit.Route(chart, "openbao", "openbao", "openbao", 8200)

	// Scraping /v1/sys/metrics relies on unauthenticated_metrics_access above.
	cdk8s.NewApiObject(chart, jsii.String("openbao-servicemonitor"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("openbao")},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{
			"matchLabels": map[string]any{"app.kubernetes.io/name": "openbao"},
		},
		"namespaceSelector": map[string]any{
			"matchNames": []string{namespace},
		},
		"endpoints": []map[string]any{
			{
				"port":     "http",
				"path":     "/v1/sys/metrics",
				"params":   map[string]any{"format": []string{"prometheus"}},
				"interval": "30s",
			},
		},
	}))

	return chart
}
