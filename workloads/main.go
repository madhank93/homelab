package main

import (
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/ai"
	"github.com/madhank93/homelab/workloads/automation"
	"github.com/madhank93/homelab/workloads/databases"
	"github.com/madhank93/homelab/workloads/hardware"
	"github.com/madhank93/homelab/workloads/management"
	"github.com/madhank93/homelab/workloads/monitoring"
	"github.com/madhank93/homelab/workloads/networking"
	"github.com/madhank93/homelab/workloads/observability"
	"github.com/madhank93/homelab/workloads/registry"
	"github.com/madhank93/homelab/workloads/secrets"
	"github.com/madhank93/homelab/workloads/security"
	"github.com/madhank93/homelab/workloads/storage"
	"github.com/madhank93/homelab/workloads/support"
)

func main() {
	rootFolder := "../app"

	// 1. Longhorn Storage
	longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/longhorn", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
	longhornApp.Synth()

	// OpenBao (secrets store)
	openBaoApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/openbao", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	secrets.NewOpenBaoChart(openBaoApp, "openbao-app", "openbao")
	openBaoApp.Synth()

	// Secrets Store CSI Driver — mounts OpenBao secrets as files in pods
	csiDriverApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/csi-driver", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	secrets.NewCsiDriverChart(csiDriverApp, "csi-driver-app")
	csiDriverApp.Synth()

	// Monitoring Stack
	grafanaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/grafana", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewGrafanaChart(grafanaApp, "grafana-app", "grafana")
	grafanaApp.Synth()

	victoriaMetricsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-metrics", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewVictoriaMetricsChart(victoriaMetricsApp, "victoria-metrics-app", "victoria-metrics")
	victoriaMetricsApp.Synth()

	victoriaLogsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-logs", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewVictoriaLogsChart(victoriaLogsApp, "victoria-logs-app", "victoria-logs")
	victoriaLogsApp.Synth()

	// Container Registry
	harborApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/harbor", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	registry.NewHarborChart(harborApp, "harbor-app", "harbor")
	harborApp.Synth()

	// CloudNativePG Operator — cluster-scoped PostgreSQL operator (dependency for n8n)
	cnpgApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/cnpg-operator", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	databases.NewCnpgOperatorChart(cnpgApp, "cnpg-operator-app")
	cnpgApp.Synth()

	// Automation
	n8nApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/n8n", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	automation.NewN8nChart(n8nApp, "n8n-app", "n8n")
	n8nApp.Synth()

	// AI/ML
	nvidiaGpuOperatorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/nvidia-gpu-operator", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	hardware.NewNvidiaDevicePluginChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator", "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()

	ollamaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/ollama", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewOllamaChart(ollamaApp, "ollama-app", "ollama")
	ollamaApp.Synth()

	comfyuiApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/comfyui", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewComfyUIChart(comfyuiApp, "comfyui-app", "comfyui")
	comfyuiApp.Synth()

	// Kubeflow — Application CR (points to workloads/ai/kubeflow/ kustomize overlay) + HTTPRoute
	kubeflowApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/kubeflow", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewKubeflowChart(kubeflowApp, "kubeflow-app", "kubeflow")
	kubeflowApp.Synth()

	// Security & Compliance
	trivyApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/trivy", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	security.NewTrivyChart(trivyApp, "trivy-app", "trivy")
	trivyApp.Synth()

	falcoApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/falco", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	security.NewFalcoChart(falcoApp, "falco-app", "falco")
	falcoApp.Synth()

	kyvernoApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/kyverno", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	security.NewKyvernoChart(kyvernoApp, "kyverno-app", "kyverno")
	kyvernoApp.Synth()

	// Observability
	otelApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/opentelemetry", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewOtelCollectorChart(otelApp, "otel-app", "opentelemetry")
	otelApp.Synth()

	// ArgoCD ServiceMonitors — ArgoCD is deployed via Pulumi; this chart only adds
	// ServiceMonitors in the argocd namespace for VMAgent to discover.
	argoCDMonitorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/argocd-monitor", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewArgoCDMonitorChart(argoCDMonitorApp, "argocd-monitor-app")
	argoCDMonitorApp.Synth()

	// NetBird routing peer — advertises 192.168.1.0/24 into WireGuard mesh
	netbirdApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/netbird", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	networking.NewNetbirdPeerChart(netbirdApp, "netbird-app", "netbird")
	netbirdApp.Synth()

	// Management Tools
	headlampApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/headlamp", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewHeadlampChart(headlampApp, "headlamp-app", "headlamp")
	headlampApp.Synth()

	rancherApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/rancher", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewRancherChart(rancherApp, "rancher-app", "cattle-system")
	rancherApp.Synth()

	// Stakater Reloader
	reloaderApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/reloader", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	support.NewReloaderChart(reloaderApp, "reloader-app", "reloader")
	reloaderApp.Synth()

	// Metrics Server — exposes pod/node CPU+memory metrics (required by Headlamp, HPA)
	metricsServerApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/metrics-server", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewMetricsServerChart(metricsServerApp, "metrics-server-app")
	metricsServerApp.Synth()
}
