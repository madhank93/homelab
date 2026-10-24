package grafana


type GrafanaGoMemLimit struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Enable automatic GOMEMLIMIT injection.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Fraction of resources.limits.memory to use as GOMEMLIMIT (e.g. 0.9 = 90 %).
	Factor *float64 `field:"optional" json:"factor" yaml:"factor"`
}

