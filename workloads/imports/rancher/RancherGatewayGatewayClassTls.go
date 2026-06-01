package rancher


// The default rancher gateway class tls configuration.
type RancherGatewayGatewayClassTls struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	SecretName *string `field:"optional" json:"secretName" yaml:"secretName"`
	Source RancherGatewayGatewayClassTlsSource `field:"optional" json:"source" yaml:"source"`
}

