package grafana


type GrafanaInitChownData struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Image *GrafanaInitChownDataImage `field:"optional" json:"image" yaml:"image"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
}

