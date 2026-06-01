package n8n


type N8NWorkflowHistory struct {
	// Whether to save workflow history versions.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Time (in hours) to keep workflow history versions for.
	//
	// To disable it, use -1 as a value.
	PruneTime *float64 `field:"required" json:"pruneTime" yaml:"pruneTime"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

