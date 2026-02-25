package n8n


// Placeholder for future Redis TLS certificates.
type N8NExternalRedisTls struct {
	// Whether to enable TLS on Redis connections.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

