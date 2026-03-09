package openbao


type OpenbaoServerTelemetryPrometheusRules struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Rules *[]interface{} `field:"optional" json:"rules" yaml:"rules"`
	Selectors interface{} `field:"optional" json:"selectors" yaml:"selectors"`
}

