package openbao


type OpenbaoServerRoute struct {
	ActiveService *bool `field:"optional" json:"activeService" yaml:"activeService"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Host *string `field:"optional" json:"host" yaml:"host"`
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	Tls interface{} `field:"optional" json:"tls" yaml:"tls"`
}

