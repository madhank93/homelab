package headlamp


// Bundled (static) plugins shipped in the Headlamp image, such as the Prometheus plugin.
type HeadlampConfigStaticPlugins struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Serve the bundled static plugins.
	//
	// Set to false to disable them (e.g. the "Show Prometheus metrics" button)
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
}

