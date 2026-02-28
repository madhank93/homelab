package n8n


// n8n license configurations.
type N8NLicense struct {
	// Activation key to initialize license.
	//
	// Not applicable if the n8n instance was already activated. For more information please refer to the following link: https://docs.n8n.io/enterprise-key/
	ActivationKey *string `field:"required" json:"activationKey" yaml:"activationKey"`
	// The auto new license configuration.
	AutoNenew *N8NLicenseAutoNenew `field:"required" json:"autoNenew" yaml:"autoNenew"`
	// Whether to enable the license.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// The name of an existing secret with license activation key.
	//
	// The secret must contain a key with the name N8N_LICENSE_ACTIVATION_KEY.
	ExistingActivationKeySecret *string `field:"required" json:"existingActivationKeySecret" yaml:"existingActivationKeySecret"`
	// Server URL to retrieve license.
	//
	// The default value is https://license.n8n.io/v1.
	ServerUrl *string `field:"required" json:"serverUrl" yaml:"serverUrl"`
	// Tenant ID associated with the license.
	//
	// Only set this variable if explicitly instructed by n8n.
	TenantId *float64 `field:"required" json:"tenantId" yaml:"tenantId"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

