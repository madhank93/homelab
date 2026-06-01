package n8n


type N8NWorkerReadinessProbeHttpGet struct {
	Path *string `field:"required" json:"path" yaml:"path"`
	Port *string `field:"required" json:"port" yaml:"port"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

