package n8n


type N8NWorkerAutoscaling struct {
	Behavior *N8NWorkerAutoscalingBehavior `field:"required" json:"behavior" yaml:"behavior"`
	// If true, the number of workers will be automatically scaled based on default metrics.
	//
	// On default, it will scale based on CPU and memory. For more information can be found here: https://kubernetes.io/docs/concepts/workloads/autoscaling/
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Maximum number of workers.
	MaxReplicas *float64 `field:"required" json:"maxReplicas" yaml:"maxReplicas"`
	Metrics *[]*N8NWorkerAutoscalingMetrics `field:"required" json:"metrics" yaml:"metrics"`
	// Minimum number of workers.
	MinReplicas *float64 `field:"required" json:"minReplicas" yaml:"minReplicas"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

