package n8n


type N8NWebhookAutoscalingBehaviorScaleDownPolicies struct {
	// Time in seconds over which the policy is evaluated.
	PeriodSeconds *float64 `field:"required" json:"periodSeconds" yaml:"periodSeconds"`
	// Type of scaling policy: Pods (fixed number) or Percent (percentage of current replicas).
	Type N8NWebhookAutoscalingBehaviorScaleDownPoliciesType `field:"required" json:"type" yaml:"type"`
	// The value to scale by (number of pods or percentage).
	Value *float64 `field:"required" json:"value" yaml:"value"`
}

