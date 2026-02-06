package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/rancher"
)

func NewRancherChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{})

	values := &rancher.RancherValues{
		AgentTlsMode: rancher.RancherAgentTlsMode_SYSTEM_HYPHEN_STORE,
		AuditLog: &rancher.RancherAuditLog{
			Level:       rancher.RancherAuditLogLevel_VALUE_0,
			Destination: rancher.RancherAuditLogDestination_SIDECAR,
		},
		Ingress: &rancher.RancherIngress{
			AdditionalValues: &map[string]any{
				"tls": map[string]any{
					"source": "letsEncrypt",
				},
			},
		},
		Service: &rancher.RancherService{
			Type:        rancher.RancherServiceType_CLUSTER_IP,
			DisableHttp: jsii.Bool(false),
		},
		AdditionalValues: &map[string]any{
			"hostname": "rancher.local",
			"letsEncrypt": map[string]any{
				"email":       "admin@example.com",
				"environment": "staging",
			},
			"bootstrapPassword": "admin",
			"replicas":          3,
			"resources": map[string]any{
				"limits": map[string]any{
					"memory": "2Gi",
					"cpu":    "1000m",
				},
				"requests": map[string]any{
					"memory": "1Gi",
					"cpu":    "500m",
				},
			},
			"antiAffinity": "preferred",
		},
	}
	rancher.NewRancher(chart, jsii.String("rancher-release"), &rancher.RancherProps{
		ReleaseName: jsii.String("rancher"),
		Values:      values,
	})
	return chart
}
