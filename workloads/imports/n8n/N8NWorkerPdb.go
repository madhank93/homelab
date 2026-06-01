package n8n


// PodDisruptionBudget configuration for the worker node.
type N8NWorkerPdb struct {
	// Whether to enable the PodDisruptionBudget for the worker node.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Maximum number of unavailable replicas for the worker node.
	MaxUnavailable interface{} `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
	// Minimum number of available replicas for the worker node.
	MinAvailable interface{} `field:"optional" json:"minAvailable" yaml:"minAvailable"`
}

