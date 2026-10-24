package grafana


type GrafanaVerticalPodAutoscaler struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ControlledResources *GrafanaVerticalPodAutoscalerControlledResources `field:"optional" json:"controlledResources" yaml:"controlledResources"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	MaxAllowed *map[string]interface{} `field:"optional" json:"maxAllowed" yaml:"maxAllowed"`
	MinAllowed *map[string]interface{} `field:"optional" json:"minAllowed" yaml:"minAllowed"`
	UpdateMode *string `field:"optional" json:"updateMode" yaml:"updateMode"`
}

