package certmanager


type HelmValuesStartupapicheckServiceAccount struct {
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Create *bool `field:"optional" json:"create" yaml:"create"`
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	Name *string `field:"optional" json:"name" yaml:"name"`
}

