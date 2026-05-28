package rancher


// The default rancher gateway configuration.
type RancherGateway struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// The default rancher gateway class configuration.
	GatewayClass *RancherGatewayGatewayClass `field:"optional" json:"gatewayClass" yaml:"gatewayClass"`
}

