package openbao


type OpenbaoCsiDaemonSetSecurityContext struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Container interface{} `field:"optional" json:"container" yaml:"container"`
	Pod interface{} `field:"optional" json:"pod" yaml:"pod"`
}

