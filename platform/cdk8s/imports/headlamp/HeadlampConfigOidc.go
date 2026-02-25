package headlamp


// OIDC configuration.
type HeadlampConfigOidc struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Issuer of the OIDC provider.
	ClientId *string `field:"optional" json:"clientId" yaml:"clientId"`
	// Client ID of the OIDC provider.
	ClientSecret *string `field:"optional" json:"clientSecret" yaml:"clientSecret"`
	// External secret to use for OIDC configuration.
	ExternalSecret *HeadlampConfigOidcExternalSecret `field:"optional" json:"externalSecret" yaml:"externalSecret"`
	// Client secret of the OIDC provider.
	IssuerUrl *string `field:"optional" json:"issuerUrl" yaml:"issuerUrl"`
	// Scopes of the OIDC provider.
	Scopes *string `field:"optional" json:"scopes" yaml:"scopes"`
	// Secret created by Headlamp to authenticate with the OIDC provider.
	Secret *HeadlampConfigOidcSecret `field:"optional" json:"secret" yaml:"secret"`
	// Use PKCE (Proof Key for Code Exchange) for enhanced security in OIDC flow.
	UsePkce *bool `field:"optional" json:"usePkce" yaml:"usePkce"`
}

