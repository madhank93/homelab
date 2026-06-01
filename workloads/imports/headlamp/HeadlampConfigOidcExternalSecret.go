package headlamp


// External secret to use for OIDC configuration.
type HeadlampConfigOidcExternalSecret struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Enable the external secret.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Name of the external secret.
	Name *string `field:"optional" json:"name" yaml:"name"`
}

