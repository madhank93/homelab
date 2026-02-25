package n8n


// Configuration for private npm registry.
type N8NNpmRegistry struct {
	// Custom .npmrc content (optional, overrides secret if provided).
	CustomNpmrc *string `field:"required" json:"customNpmrc" yaml:"customNpmrc"`
	// Enable private npm registry.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Key in the secret for the .npmrc content or auth token.
	SecretKey *string `field:"required" json:"secretKey" yaml:"secretKey"`
	// Name of the Kubernetes secret containing npm registry credentials.
	SecretName *string `field:"required" json:"secretName" yaml:"secretName"`
	// URL of the private npm registry (e.g., https://registry.npmjs.org/).
	Url *string `field:"required" json:"url" yaml:"url"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

