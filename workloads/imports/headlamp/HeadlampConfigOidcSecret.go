package headlamp


// Secret created by Headlamp to authenticate with the OIDC provider.
type HeadlampConfigOidcSecret struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Create the secret.
	Create *bool `field:"optional" json:"create" yaml:"create"`
	// Name of the secret.
	Name *string `field:"optional" json:"name" yaml:"name"`
}

