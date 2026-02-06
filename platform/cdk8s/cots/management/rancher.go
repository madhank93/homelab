package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewRancherChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]interface{}{
		"agentTLSMode": "system-store",
		"auditLog": map[string]interface{}{
			"level":       0,
			"destination": "sidecar",
		},
		"ingress": map[string]interface{}{
			"extraValues": map[string]interface{}{
				"tls": map[string]interface{}{
					"source": "secret",
				},
			},
		},
		"service": map[string]interface{}{
			"type":        "ClusterIP",
			"disableHttp": false,
		},
		"hostname":          "rancher.local",
		"bootstrapPassword": "admin",
		"replicas":          3,
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
			"requests": map[string]interface{}{
				"memory": "1Gi",
				"cpu":    "500m",
			},
		},
		"antiAffinity": "preferred",
	}

	cdk8s.NewHelm(chart, jsii.String("rancher-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("/Volumes/work/git-repos/homelab/platform/cdk8s/imports/charts/rancher"),
		ReleaseName: jsii.String("rancher"),
		Values:      &values,
	})
	return chart
}
