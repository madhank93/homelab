package grafana


type GrafanaTestFrameworkImage struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// The Docker registry.
	Registry *string `field:"optional" json:"registry" yaml:"registry"`
	Repository *string `field:"optional" json:"repository" yaml:"repository"`
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
}

