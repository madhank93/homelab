package n8n


// Metrics and labels to include in the ServiceMonitor.
type N8NServiceMonitorInclude struct {
	// Whether to include api endpoints.
	ApiEndpoints *bool `field:"required" json:"apiEndpoints" yaml:"apiEndpoints"`
	// Whether to include api method label.
	ApiMethodLabel *bool `field:"required" json:"apiMethodLabel" yaml:"apiMethodLabel"`
	// Whether to include api path label.
	ApiPathLabel *bool `field:"required" json:"apiPathLabel" yaml:"apiPathLabel"`
	// Whether to include api status code label.
	ApiStatusCodeLabel *bool `field:"required" json:"apiStatusCodeLabel" yaml:"apiStatusCodeLabel"`
	// Whether to include cache metrics.
	CacheMetrics *bool `field:"required" json:"cacheMetrics" yaml:"cacheMetrics"`
	// Whether to include credential type label.
	CredentialTypeLabel *bool `field:"required" json:"credentialTypeLabel" yaml:"credentialTypeLabel"`
	// Whether to include default metrics.
	DefaultMetrics *bool `field:"required" json:"defaultMetrics" yaml:"defaultMetrics"`
	// Whether to include message event bus metrics.
	MessageEventBusMetrics *bool `field:"required" json:"messageEventBusMetrics" yaml:"messageEventBusMetrics"`
	// Whether to include node type label.
	NodeTypeLabel *bool `field:"required" json:"nodeTypeLabel" yaml:"nodeTypeLabel"`
	// Whether to include queue metrics.
	QueueMetrics *bool `field:"required" json:"queueMetrics" yaml:"queueMetrics"`
	// Whether to include workflow id label.
	WorkflowIdLabel *bool `field:"required" json:"workflowIdLabel" yaml:"workflowIdLabel"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

