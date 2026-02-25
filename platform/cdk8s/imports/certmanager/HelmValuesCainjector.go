package certmanager


type HelmValuesCainjector struct {
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	DeploymentAnnotations interface{} `field:"optional" json:"deploymentAnnotations" yaml:"deploymentAnnotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	ExtraEnv *[]interface{} `field:"optional" json:"extraEnv" yaml:"extraEnv"`
	FeatureGates *string `field:"optional" json:"featureGates" yaml:"featureGates"`
	Image *HelmValuesCainjectorImage `field:"optional" json:"image" yaml:"image"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PodAnnotations interface{} `field:"optional" json:"podAnnotations" yaml:"podAnnotations"`
	PodDisruptionBudget *HelmValuesCainjectorPodDisruptionBudget `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	PodLabels interface{} `field:"optional" json:"podLabels" yaml:"podLabels"`
	ReplicaCount *float64 `field:"optional" json:"replicaCount" yaml:"replicaCount"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount *HelmValuesCainjectorServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ServiceAnnotations interface{} `field:"optional" json:"serviceAnnotations" yaml:"serviceAnnotations"`
	ServiceLabels interface{} `field:"optional" json:"serviceLabels" yaml:"serviceLabels"`
	Strategy interface{} `field:"optional" json:"strategy" yaml:"strategy"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints *[]interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

