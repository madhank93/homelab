package grafana


type GrafanaImageRendererAutoscaling struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Behavior interface{} `field:"optional" json:"behavior" yaml:"behavior"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	MaxReplicas *float64 `field:"optional" json:"maxReplicas" yaml:"maxReplicas"`
	MinReplicas *float64 `field:"optional" json:"minReplicas" yaml:"minReplicas"`
	TargetCpu interface{} `field:"optional" json:"targetCpu" yaml:"targetCpu"`
	TargetMemory interface{} `field:"optional" json:"targetMemory" yaml:"targetMemory"`
}

