package n8n


// Policy to select among multiple policies: Max (default), Min, or Disabled.
type N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy string

const (
	// Max.
	N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_MAX N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy = "MAX"
	// Min.
	N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_MIN N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy = "MIN"
	// Disabled.
	N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_DISABLED N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy = "DISABLED"
)

