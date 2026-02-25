package n8n


// Type of scaling policy: Pods (fixed number) or Percent (percentage of current replicas).
type N8NWebhookAutoscalingBehaviorScaleUpPoliciesType string

const (
	// Pods.
	N8NWebhookAutoscalingBehaviorScaleUpPoliciesType_PODS N8NWebhookAutoscalingBehaviorScaleUpPoliciesType = "PODS"
	// Percent.
	N8NWebhookAutoscalingBehaviorScaleUpPoliciesType_PERCENT N8NWebhookAutoscalingBehaviorScaleUpPoliciesType = "PERCENT"
)

