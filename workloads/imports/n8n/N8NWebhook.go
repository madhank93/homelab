package n8n


type N8NWebhook struct {
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity.
	Affinity interface{} `field:"required" json:"affinity" yaml:"affinity"`
	// If true, all k8s nodes will deploy exatly one worker pod.
	AllNodes *bool `field:"required" json:"allNodes" yaml:"allNodes"`
	Autoscaling *N8NWebhookAutoscaling `field:"required" json:"autoscaling" yaml:"autoscaling"`
	// number of webhooks.
	Count *float64 `field:"required" json:"count" yaml:"count"`
	// Additional containers for the webhook pod.
	ExtraContainers *[]interface{} `field:"required" json:"extraContainers" yaml:"extraContainers"`
	// Extra environment variables.
	ExtraEnvVars interface{} `field:"required" json:"extraEnvVars" yaml:"extraEnvVars"`
	// Extra secrets for environment variables.
	ExtraSecretNamesForEnvFrom *[]interface{} `field:"required" json:"extraSecretNamesForEnvFrom" yaml:"extraSecretNamesForEnvFrom"`
	// Host aliases for the webhook pod.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/#adding-additional-entries-with-hostaliases
	HostAliases *[]*N8NWebhookHostAliases `field:"required" json:"hostAliases" yaml:"hostAliases"`
	// Additional init containers for the pod.
	InitContainers *[]interface{} `field:"required" json:"initContainers" yaml:"initContainers"`
	// This is to setup the liveness probe for the webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	LivenessProbe *N8NWebhookLivenessProbe `field:"required" json:"livenessProbe" yaml:"livenessProbe"`
	// MCP webhook configuration.
	//
	// This is only used when the webhook mode is set to `queue` and the database type is set to `postgresdb`.
	Mcp *N8NWebhookMcp `field:"required" json:"mcp" yaml:"mcp"`
	// Use `regular` to use main node as webhook node, or use `queue` to have webhook nodes.
	Mode N8NWebhookMode `field:"required" json:"mode" yaml:"mode"`
	// PodDisruptionBudget configuration for the worker node.
	Pdb *N8NWebhookPdb `field:"required" json:"pdb" yaml:"pdb"`
	// This is to setup the readiness probe for the webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	ReadinessProbe *N8NWebhookReadinessProbe `field:"required" json:"readinessProbe" yaml:"readinessProbe"`
	// This block is for setting up the resource management for the pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Resources *N8NWebhookResources `field:"required" json:"resources" yaml:"resources"`
	// Runtime class name for the webhook pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/containers/runtime-class/
	RuntimeClassName *string `field:"required" json:"runtimeClassName" yaml:"runtimeClassName"`
	// This is to setup the startup probe for the webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	StartupProbe *N8NWebhookStartupProbe `field:"required" json:"startupProbe" yaml:"startupProbe"`
	// Webhook url together with http schema.
	Url *string `field:"required" json:"url" yaml:"url"`
	// Additional volumeMounts on the output Deployment definition.
	VolumeMounts *[]interface{} `field:"required" json:"volumeMounts" yaml:"volumeMounts"`
	// Additional volumes on the output Deployment definition.
	Volumes *[]interface{} `field:"required" json:"volumes" yaml:"volumes"`
	// This is to setup the wait for the main node to be ready.
	WaitMainNodeReady *N8NWebhookWaitMainNodeReady `field:"required" json:"waitMainNodeReady" yaml:"waitMainNodeReady"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

