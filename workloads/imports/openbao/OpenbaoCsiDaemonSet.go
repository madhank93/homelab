package openbao


type OpenbaoCsiDaemonSet struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	KubeletRootDir *string `field:"optional" json:"kubeletRootDir" yaml:"kubeletRootDir"`
	ProvidersDir *string `field:"optional" json:"providersDir" yaml:"providersDir"`
	SecurityContext *OpenbaoCsiDaemonSetSecurityContext `field:"optional" json:"securityContext" yaml:"securityContext"`
	UpdateStrategy *OpenbaoCsiDaemonSetUpdateStrategy `field:"optional" json:"updateStrategy" yaml:"updateStrategy"`
}

