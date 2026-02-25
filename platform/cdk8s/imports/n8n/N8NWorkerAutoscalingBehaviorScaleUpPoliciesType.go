package n8n


// Type of scaling policy: Pods (fixed number) or Percent (percentage of current replicas).
type N8NWorkerAutoscalingBehaviorScaleUpPoliciesType string

const (
	// Pods.
	N8NWorkerAutoscalingBehaviorScaleUpPoliciesType_PODS N8NWorkerAutoscalingBehaviorScaleUpPoliciesType = "PODS"
	// Percent.
	N8NWorkerAutoscalingBehaviorScaleUpPoliciesType_PERCENT N8NWorkerAutoscalingBehaviorScaleUpPoliciesType = "PERCENT"
)

