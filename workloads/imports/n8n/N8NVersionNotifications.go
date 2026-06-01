package n8n


type N8NVersionNotifications struct {
	// Whether to request notifications about new n8n versions.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Endpoint to retrieve n8n version information from.
	Endpoint *string `field:"required" json:"endpoint" yaml:"endpoint"`
	// URL for versions panel to page instructing user on how to update n8n instance.
	InfoUrl *string `field:"required" json:"infoUrl" yaml:"infoUrl"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

