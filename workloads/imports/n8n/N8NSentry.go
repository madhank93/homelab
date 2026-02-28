package n8n


type N8NSentry struct {
	// Sentry DSN for backend.
	BackendDsn *string `field:"required" json:"backendDsn" yaml:"backendDsn"`
	// Whether sentry is enabled.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Sentry DSN for external task runners.
	ExternalTaskRunnersDsn *string `field:"required" json:"externalTaskRunnersDsn" yaml:"externalTaskRunnersDsn"`
	// Sentry DSN for frontend.
	FrontendDsn *string `field:"required" json:"frontendDsn" yaml:"frontendDsn"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

