package n8n


type N8NWorkerAutoscalingBehavior struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ScaleDown *N8NWorkerAutoscalingBehaviorScaleDown `field:"optional" json:"scaleDown" yaml:"scaleDown"`
	ScaleUp *N8NWorkerAutoscalingBehaviorScaleUp `field:"optional" json:"scaleUp" yaml:"scaleUp"`
}

