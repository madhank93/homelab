package headlamp


type HeadlampPodDisruptionBudget struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Enable PodDisruptionBudget.
	//
	// See: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Maximum number/percentage of pods that may be made unavailable.
	MaxUnavailable interface{} `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
	// Minimum number/percentage of pods that should remain scheduled.
	//
	// When it's set, maxUnavailable must be disabled by `maxUnavailable: null`.
	MinAvailable interface{} `field:"optional" json:"minAvailable" yaml:"minAvailable"`
	// How are unhealthy, but running, pods counted for eviction.
	UnhealthyPodEvictionPolicy interface{} `field:"optional" json:"unhealthyPodEvictionPolicy" yaml:"unhealthyPodEvictionPolicy"`
}

