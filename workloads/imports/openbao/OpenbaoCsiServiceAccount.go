package openbao


type OpenbaoCsiServiceAccount struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
}

