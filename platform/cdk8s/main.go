package main

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/cots"
)

func main() {
	// Cert-Manager
	certMgrApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String("app/cert-manager-app"),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	cots.NewCertManagerChart(certMgrApp, "cert-manager-app")
	certMgrApp.Synth()

	// NVIDIA GPU Operator
	nvidiaGpuOperatorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String("app/nvidia-gpu-operator"),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	cots.NewNvidiaGpuOperatorChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()
}
