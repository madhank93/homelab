package n8n


type N8NServiceMonitorMetricRelabelings struct {
	// The relabeling action to perform.
	Action N8NServiceMonitorMetricRelabelingsAction `field:"required" json:"action" yaml:"action"`
	// Modulus to use with hashmod action.
	Modulus *float64 `field:"optional" json:"modulus" yaml:"modulus"`
	// The regular expression to match against source labels.
	Regex *string `field:"optional" json:"regex" yaml:"regex"`
	// Replacement value for the 'replace' action.
	Replacement *string `field:"optional" json:"replacement" yaml:"replacement"`
	// Separator used when concatenating source labels.
	Separator *string `field:"optional" json:"separator" yaml:"separator"`
	// The source labels to relabel from.
	SourceLabels *[]*string `field:"optional" json:"sourceLabels" yaml:"sourceLabels"`
	// The label to write the result to.
	TargetLabel *string `field:"optional" json:"targetLabel" yaml:"targetLabel"`
}

