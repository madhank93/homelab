package n8n


// Type of scaling policy: Pods (fixed number) or Percent (percentage of current replicas).
type N8NWorkerAutoscalingBehaviorScaleDownPoliciesType string

const (
	// Pods.
	N8NWorkerAutoscalingBehaviorScaleDownPoliciesType_PODS N8NWorkerAutoscalingBehaviorScaleDownPoliciesType = "PODS"
	// Percent.
	N8NWorkerAutoscalingBehaviorScaleDownPoliciesType_PERCENT N8NWorkerAutoscalingBehaviorScaleDownPoliciesType = "PERCENT"
)

