package n8n


// Policy to select among multiple policies: Max (default), Min, or Disabled.
type N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy string

const (
	// Max.
	N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_MAX N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy = "MAX"
	// Min.
	N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_MIN N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy = "MIN"
	// Disabled.
	N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_DISABLED N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy = "DISABLED"
)

