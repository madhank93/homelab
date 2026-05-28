package rancher


// The default rancher gateway class configuration.
type RancherGatewayGatewayClass struct {
	AdditionalListeners *[]*RancherGatewayGatewayClassAdditionalListeners `field:"optional" json:"additionalListeners" yaml:"additionalListeners"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The default rancher gateway class ports configuration.
	Ports *RancherGatewayGatewayClassPorts `field:"optional" json:"ports" yaml:"ports"`
	// The default rancher gateway class tls configuration.
	Tls *RancherGatewayGatewayClassTls `field:"optional" json:"tls" yaml:"tls"`
}

