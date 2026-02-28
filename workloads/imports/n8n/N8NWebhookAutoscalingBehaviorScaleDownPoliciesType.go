package n8n


// Type of scaling policy: Pods (fixed number) or Percent (percentage of current replicas).
type N8NWebhookAutoscalingBehaviorScaleDownPoliciesType string

const (
	// Pods.
	N8NWebhookAutoscalingBehaviorScaleDownPoliciesType_PODS N8NWebhookAutoscalingBehaviorScaleDownPoliciesType = "PODS"
	// Percent.
	N8NWebhookAutoscalingBehaviorScaleDownPoliciesType_PERCENT N8NWebhookAutoscalingBehaviorScaleDownPoliciesType = "PERCENT"
)

