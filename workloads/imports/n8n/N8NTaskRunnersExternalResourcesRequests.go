package n8n


type N8NTaskRunnersExternalResourcesRequests struct {
	Cpu *string `field:"required" json:"cpu" yaml:"cpu"`
	Memory *string `field:"required" json:"memory" yaml:"memory"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

