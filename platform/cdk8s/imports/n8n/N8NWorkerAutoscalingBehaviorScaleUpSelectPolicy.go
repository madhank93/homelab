package n8n


// Policy to select among multiple policies: Max (default), Min, or Disabled.
type N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy string

const (
	// Max.
	N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_MAX N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy = "MAX"
	// Min.
	N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_MIN N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy = "MIN"
	// Disabled.
	N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_DISABLED N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy = "DISABLED"
)

