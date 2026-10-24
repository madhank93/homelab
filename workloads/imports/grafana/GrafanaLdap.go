package grafana


type GrafanaLdap struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Config *string `field:"optional" json:"config" yaml:"config"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExistingSecret *string `field:"optional" json:"existingSecret" yaml:"existingSecret"`
}

