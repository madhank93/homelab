package n8n


// This is to setup the wait for the main node to be ready.
type N8NWebhookWaitMainNodeReady struct {
	// The additional parameters to use part of wget command.
	//
	// e.g. --no-check-certificate
	AdditionalParameters *[]interface{} `field:"required" json:"additionalParameters" yaml:"additionalParameters"`
	// Whether to enable the wait for the main node to be ready.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// The health check path to use for request to the main node.
	HealthCheckPath *string `field:"required" json:"healthCheckPath" yaml:"healthCheckPath"`
	// The schema to use for request to the main node.
	//
	// On default, it will use identify the schema from the main N8N_PROTOCOL environment variable or use http.
	OverwriteSchema *string `field:"required" json:"overwriteSchema" yaml:"overwriteSchema"`
	// The URL to use for request to the main node.
	//
	// On default, it will use service name and port.
	OverwriteUrl *string `field:"required" json:"overwriteUrl" yaml:"overwriteUrl"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

