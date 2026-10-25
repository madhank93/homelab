package main

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/ai"
	"github.com/madhank93/homelab/workloads/automation"
	"github.com/madhank93/homelab/workloads/databases"
	"github.com/madhank93/homelab/workloads/gaming"
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

// Each entry synthesizes into ../app/<dir>, which Argo CD deploys as one Application.
var apps = []struct {
	dir   string
	build func(app cdk8s.App)
}{
	{"longhorn", func(a cdk8s.App) { storage.NewLonghornChart(a, "longhorn-app", "longhorn-system") }},
	{"openbao", func(a cdk8s.App) { secrets.NewOpenBaoChart(a, "openbao-app", "openbao") }},
	{"csi-driver", func(a cdk8s.App) { secrets.NewCsiDriverChart(a, "csi-driver-app") }},
	{"grafana", func(a cdk8s.App) { monitoring.NewGrafanaChart(a, "grafana-app", "grafana") }},
	{"victoria-metrics", func(a cdk8s.App) {
		observability.NewVictoriaMetricsChart(a, "victoria-metrics-app", "victoria-metrics")
	}},
	{"victoria-logs", func(a cdk8s.App) { observability.NewVictoriaLogsChart(a, "victoria-logs-app", "victoria-logs") }},
	{"harbor", func(a cdk8s.App) { registry.NewHarborChart(a, "harbor-app", "harbor") }},
	{"cnpg-operator", func(a cdk8s.App) { databases.NewCnpgOperatorChart(a, "cnpg-operator-app") }},
	{"n8n", func(a cdk8s.App) { automation.NewN8nChart(a, "n8n-app", "n8n") }},
	{"nvidia-gpu-operator", func(a cdk8s.App) {
		hardware.NewNvidiaDevicePluginChart(a, "nvidia-gpu-operator", "nvidia-gpu-operator")
	}},
	{"ollama", func(a cdk8s.App) { ai.NewOllamaChart(a, "ollama-app", "ollama") }},
	{"comfyui", func(a cdk8s.App) { ai.NewComfyUIChart(a, "comfyui-app", "comfyui") }},
	// The Kubeflow platform itself is a pre-built kustomize overlay (workloads/ai/kubeflow/).
	{"kubeflow", func(a cdk8s.App) {
		ai.NewKubeflowChart(a, "kubeflow-app", "kubeflow")
		ai.NewNotebookGatewayControllerChart(a, "notebook-gateway-controller", "kubeflow")
	}},
	{"trivy", func(a cdk8s.App) { security.NewTrivyChart(a, "trivy-app", "trivy") }},
	{"falco", func(a cdk8s.App) { security.NewFalcoChart(a, "falco-app", "falco") }},
	{"kyverno", func(a cdk8s.App) { security.NewKyvernoChart(a, "kyverno-app", "kyverno") }},
	{"opentelemetry", func(a cdk8s.App) { observability.NewOtelCollectorChart(a, "otel-app", "opentelemetry") }},
	// Argo CD itself is deployed by Pulumi; this adds only its ServiceMonitors.
	{"argocd-monitor", func(a cdk8s.App) { observability.NewArgoCDMonitorChart(a, "argocd-monitor-app") }},
	{"netbird", func(a cdk8s.App) { networking.NewNetbirdPeerChart(a, "netbird-app", "netbird") }},
	{"headlamp", func(a cdk8s.App) { management.NewHeadlampChart(a, "headlamp-app", "headlamp") }},
	{"reloader", func(a cdk8s.App) { support.NewReloaderChart(a, "reloader-app", "reloader") }},
	{"cs2", func(a cdk8s.App) { gaming.NewCS2Chart(a, "cs2-app", "cs2") }},
	{"metrics-server", func(a cdk8s.App) { monitoring.NewMetricsServerChart(a, "metrics-server-app") }},
}

func main() {
	for _, entry := range apps {
		app := cdk8s.NewApp(&cdk8s.AppProps{
			Outdir:         jsii.String("../app/" + entry.dir),
			YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
		})
		entry.build(app)
		app.Synth()
	}
}
