package openbao


type OpenbaoServerNetworkPolicy struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Egress *[]interface{} `field:"optional" json:"egress" yaml:"egress"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Ingress *[]interface{} `field:"optional" json:"ingress" yaml:"ingress"`
}

