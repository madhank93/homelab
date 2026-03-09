package openbao


type OpenbaoInjectorAgentImage struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Repository *string `field:"optional" json:"repository" yaml:"repository"`
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
}

