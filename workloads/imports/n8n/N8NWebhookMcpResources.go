package n8n


// This block is for setting up the resource management for the mcp webhook pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
type N8NWebhookMcpResources struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Limits *N8NWebhookMcpResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	Requests *N8NWebhookMcpResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

