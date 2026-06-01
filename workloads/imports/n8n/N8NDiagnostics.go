package n8n


type N8NDiagnostics struct {
	// Diagnostics config for backend.
	BackendConfig *string `field:"required" json:"backendConfig" yaml:"backendConfig"`
	// Whether diagnostics are enabled.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Diagnostics config for frontend.
	FrontendConfig *string `field:"required" json:"frontendConfig" yaml:"frontendConfig"`
	PostHog *N8NDiagnosticsPostHog `field:"required" json:"postHog" yaml:"postHog"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

