package main

import (
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/cots/ai"
	"github.com/madhank93/homelab/cdk8s/cots/seccomp"
)

func main() {

	rootFolder := "../../app"

	// Cert-Manager
	certMgrApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/cert-manager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewCertManagerChart(certMgrApp, "cert-manager-app")
	certMgrApp.Synth()

	// NVIDIA GPU Operator
	nvidiaGpuOperatorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/nvidia-gpu-operator", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewNvidiaGpuOperatorChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()
}
