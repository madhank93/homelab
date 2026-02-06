package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/certmanager"
)

func NewCertManagerChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{})

	values := map[string]interface{}{
		"installCRDs": jsii.Bool(true),
	}

	certmanager.NewCertmanager(chart, jsii.String("app-release"), &certmanager.CertmanagerProps{
		ReleaseName: jsii.String("cert-manager"),
		Values:      &values,
	})

	return chart
}
