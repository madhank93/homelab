package openbao


type OpenbaoCsiAgent struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	Image *OpenbaoCsiAgentImage `field:"optional" json:"image" yaml:"image"`
	LogFormat *string `field:"optional" json:"logFormat" yaml:"logFormat"`
	LogLevel *string `field:"optional" json:"logLevel" yaml:"logLevel"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
}

