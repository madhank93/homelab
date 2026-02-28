package n8n


type N8NDiagnosticsPostHog struct {
	// API host for PostHog.
	ApiHost *string `field:"required" json:"apiHost" yaml:"apiHost"`
	// API key for PostHog.
	ApiKey *string `field:"required" json:"apiKey" yaml:"apiKey"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

