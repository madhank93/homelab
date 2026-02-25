package n8n


type N8NWorkerAutoscalingBehaviorScaleUp struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// List of scaling policies for scale-up.
	Policies *[]*N8NWorkerAutoscalingBehaviorScaleUpPolicies `field:"optional" json:"policies" yaml:"policies"`
	// Policy to select among multiple policies: Max (default), Min, or Disabled.
	SelectPolicy N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy `field:"optional" json:"selectPolicy" yaml:"selectPolicy"`
	// Time in seconds to wait before scaling up again (default: 0).
	StabilizationWindowSeconds *float64 `field:"optional" json:"stabilizationWindowSeconds" yaml:"stabilizationWindowSeconds"`
}

