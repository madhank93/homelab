package n8n


// MCP webhook configuration.
//
// This is only used when the webhook mode is set to `queue` and the database type is set to `postgresdb`.
type N8NWebhookMcp struct {
	// Webhook node affinity for the mcp webhook pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity
	Affinity interface{} `field:"required" json:"affinity" yaml:"affinity"`
	// Whether to enable MCP webhook.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Additional containers for the mcp webhook pod.
	ExtraContainers *[]interface{} `field:"required" json:"extraContainers" yaml:"extraContainers"`
	// Extra environment variables for the mcp webhook pod.
	ExtraEnvVars interface{} `field:"required" json:"extraEnvVars" yaml:"extraEnvVars"`
	// Extra secrets for environment variables for the mcp webhook pod.
	ExtraSecretNamesForEnvFrom *[]interface{} `field:"required" json:"extraSecretNamesForEnvFrom" yaml:"extraSecretNamesForEnvFrom"`
	// Host aliases for the mcp webhook pod.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/#adding-additional-entries-with-hostaliases
	HostAliases *[]*N8NWebhookMcpHostAliases `field:"required" json:"hostAliases" yaml:"hostAliases"`
	// Additional init containers for the mcp webhook pod.
	InitContainers *[]interface{} `field:"required" json:"initContainers" yaml:"initContainers"`
	// This is to setup the liveness probe for the mcp webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	LivenessProbe *N8NWebhookMcpLivenessProbe `field:"required" json:"livenessProbe" yaml:"livenessProbe"`
	// This is to setup the readiness probe for the mcp webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	ReadinessProbe *N8NWebhookMcpReadinessProbe `field:"required" json:"readinessProbe" yaml:"readinessProbe"`
	// This block is for setting up the resource management for the mcp webhook pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Resources *N8NWebhookMcpResources `field:"required" json:"resources" yaml:"resources"`
	// This is to setup the startup probe for the mcp webhook pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	StartupProbe *N8NWebhookMcpStartupProbe `field:"required" json:"startupProbe" yaml:"startupProbe"`
	// Additional volumeMounts on the output Deployment definition for the mcp webhook pod.
	VolumeMounts *[]interface{} `field:"required" json:"volumeMounts" yaml:"volumeMounts"`
	// Additional volumes on the output Deployment definition for the mcp webhook pod.
	Volumes *[]interface{} `field:"required" json:"volumes" yaml:"volumes"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

