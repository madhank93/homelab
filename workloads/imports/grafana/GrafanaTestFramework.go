package grafana


type GrafanaTestFramework struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Image *GrafanaTestFrameworkImage `field:"optional" json:"image" yaml:"image"`
	ImagePullPolicy *string `field:"optional" json:"imagePullPolicy" yaml:"imagePullPolicy"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
}

