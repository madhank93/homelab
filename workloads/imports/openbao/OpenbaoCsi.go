package openbao


type OpenbaoCsi struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Agent *OpenbaoCsiAgent `field:"optional" json:"agent" yaml:"agent"`
	DaemonSet *OpenbaoCsiDaemonSet `field:"optional" json:"daemonSet" yaml:"daemonSet"`
	Debug *bool `field:"optional" json:"debug" yaml:"debug"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	Image *OpenbaoCsiImage `field:"optional" json:"image" yaml:"image"`
	LivenessProbe *OpenbaoCsiLivenessProbe `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	Pod *OpenbaoCsiPod `field:"optional" json:"pod" yaml:"pod"`
	PriorityClassName *string `field:"optional" json:"priorityClassName" yaml:"priorityClassName"`
	ReadinessProbe *OpenbaoCsiReadinessProbe `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	ServiceAccount *OpenbaoCsiServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

