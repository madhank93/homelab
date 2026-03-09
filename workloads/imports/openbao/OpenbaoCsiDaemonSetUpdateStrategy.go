package openbao


type OpenbaoCsiDaemonSetUpdateStrategy struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	MaxUnavailable *string `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
	Type *string `field:"optional" json:"type" yaml:"type"`
}

