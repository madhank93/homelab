package grafana


type GrafanaImage struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	PullPolicy *string `field:"optional" json:"pullPolicy" yaml:"pullPolicy"`
	PullSecrets *[]interface{} `field:"optional" json:"pullSecrets" yaml:"pullSecrets"`
	// The Docker registry.
	Registry *string `field:"optional" json:"registry" yaml:"registry"`
	// Docker image repository.
	Repository *string `field:"optional" json:"repository" yaml:"repository"`
	Sha *string `field:"optional" json:"sha" yaml:"sha"`
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
}

