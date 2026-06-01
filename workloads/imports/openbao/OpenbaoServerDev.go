package openbao


type OpenbaoServerDev struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	DevRootToken *string `field:"optional" json:"devRootToken" yaml:"devRootToken"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
}

