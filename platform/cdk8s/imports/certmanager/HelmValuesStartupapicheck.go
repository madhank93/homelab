package certmanager


type HelmValuesStartupapicheck struct {
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	BackoffLimit *float64 `field:"optional" json:"backoffLimit" yaml:"backoffLimit"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	ExtraEnv *[]interface{} `field:"optional" json:"extraEnv" yaml:"extraEnv"`
	Image *HelmValuesStartupapicheckImage `field:"optional" json:"image" yaml:"image"`
	JobAnnotations interface{} `field:"optional" json:"jobAnnotations" yaml:"jobAnnotations"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PodAnnotations interface{} `field:"optional" json:"podAnnotations" yaml:"podAnnotations"`
	PodLabels interface{} `field:"optional" json:"podLabels" yaml:"podLabels"`
	Rbac *HelmValuesStartupapicheckRbac `field:"optional" json:"rbac" yaml:"rbac"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount *HelmValuesStartupapicheckServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	Timeout *string `field:"optional" json:"timeout" yaml:"timeout"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

