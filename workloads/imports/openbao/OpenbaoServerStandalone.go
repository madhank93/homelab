package openbao


type OpenbaoServerStandalone struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
}

