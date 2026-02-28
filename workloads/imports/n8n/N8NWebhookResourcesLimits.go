package n8n


type N8NWebhookResourcesLimits struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Cpu *string `field:"optional" json:"cpu" yaml:"cpu"`
	Memory *string `field:"optional" json:"memory" yaml:"memory"`
}

