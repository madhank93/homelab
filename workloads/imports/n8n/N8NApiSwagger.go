package n8n


// Whether to enable the Swagger UI for the Public API.
type N8NApiSwagger struct {
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

