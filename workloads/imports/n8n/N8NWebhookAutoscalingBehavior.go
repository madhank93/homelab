package n8n


type N8NWebhookAutoscalingBehavior struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ScaleDown *N8NWebhookAutoscalingBehaviorScaleDown `field:"optional" json:"scaleDown" yaml:"scaleDown"`
	ScaleUp *N8NWebhookAutoscalingBehaviorScaleUp `field:"optional" json:"scaleUp" yaml:"scaleUp"`
}

