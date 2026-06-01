package n8n


type N8NApi struct {
	// Whether to enable the Public API.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Path segment for the Public API.
	Path *string `field:"required" json:"path" yaml:"path"`
	// Whether to enable the Swagger UI for the Public API.
	Swagger *N8NApiSwagger `field:"required" json:"swagger" yaml:"swagger"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

