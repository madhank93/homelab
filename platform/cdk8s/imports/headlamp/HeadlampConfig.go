package headlamp


// Headlamp deployment configuration.
type HeadlampConfig struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Base URL of the application.
	BaseUrl *string `field:"optional" json:"baseUrl" yaml:"baseUrl"`
	// Extra arguments to pass to the application.
	ExtraArgs *[]*string `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	// OIDC configuration.
	Oidc *HeadlampConfigOidc `field:"optional" json:"oidc" yaml:"oidc"`
	// Directory to load plugins from.
	PluginsDir *string `field:"optional" json:"pluginsDir" yaml:"pluginsDir"`
	// Path of certificate file for TLS.
	TlsCertPath *string `field:"optional" json:"tlsCertPath" yaml:"tlsCertPath"`
	// Path of private key file for TLS.
	TlsKeyPath *string `field:"optional" json:"tlsKeyPath" yaml:"tlsKeyPath"`
}

