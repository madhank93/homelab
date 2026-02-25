package n8n


type N8NWebhookMcpStartupProbeExec struct {
	Command *[]interface{} `field:"required" json:"command" yaml:"command"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

