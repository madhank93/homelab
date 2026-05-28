package rancher


// The default rancher gateway class ports configuration.
type RancherGatewayGatewayClassPorts struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Http *float64 `field:"optional" json:"http" yaml:"http"`
	Https *float64 `field:"optional" json:"https" yaml:"https"`
}

