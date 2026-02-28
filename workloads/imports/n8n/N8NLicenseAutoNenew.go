package n8n


// The auto new license configuration.
type N8NLicenseAutoNenew struct {
	// Enables (true) or disables (false) autorenewal for licenses.
	//
	// If disabled, you need to manually renew the license every 10 days by navigating to Settings > Usage and plan, and pressing F5. Failure to renew the license will disable Enterprise features.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// The offset in hours before expiry a license should automatically renew.
	//
	// The default value is 72 hours (3 days).
	OffsetInHours *float64 `field:"required" json:"offsetInHours" yaml:"offsetInHours"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

