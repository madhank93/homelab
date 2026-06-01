package n8n


// Policy to select among multiple policies: Max (default), Min, or Disabled.
type N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy string

const (
	// Max.
	N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_MAX N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy = "MAX"
	// Min.
	N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_MIN N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy = "MIN"
	// Disabled.
	N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_DISABLED N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy = "DISABLED"
)

