package certmanager


type HelmValuesWebhook struct {
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	DeploymentAnnotations interface{} `field:"optional" json:"deploymentAnnotations" yaml:"deploymentAnnotations"`
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	ExtraEnv *[]interface{} `field:"optional" json:"extraEnv" yaml:"extraEnv"`
	FeatureGates *string `field:"optional" json:"featureGates" yaml:"featureGates"`
	HostNetwork *bool `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	Image *HelmValuesWebhookImage `field:"optional" json:"image" yaml:"image"`
	LivenessProbe interface{} `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	LoadBalancerIp *string `field:"optional" json:"loadBalancerIp" yaml:"loadBalancerIp"`
	MutatingWebhookConfiguration *HelmValuesWebhookMutatingWebhookConfiguration `field:"optional" json:"mutatingWebhookConfiguration" yaml:"mutatingWebhookConfiguration"`
	MutatingWebhookConfigurationAnnotations interface{} `field:"optional" json:"mutatingWebhookConfigurationAnnotations" yaml:"mutatingWebhookConfigurationAnnotations"`
	NetworkPolicy *HelmValuesWebhookNetworkPolicy `field:"optional" json:"networkPolicy" yaml:"networkPolicy"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PodAnnotations interface{} `field:"optional" json:"podAnnotations" yaml:"podAnnotations"`
	PodDisruptionBudget *HelmValuesWebhookPodDisruptionBudget `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	PodLabels interface{} `field:"optional" json:"podLabels" yaml:"podLabels"`
	ReadinessProbe interface{} `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	ReplicaCount *float64 `field:"optional" json:"replicaCount" yaml:"replicaCount"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurePort *float64 `field:"optional" json:"securePort" yaml:"securePort"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount *HelmValuesWebhookServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ServiceAnnotations interface{} `field:"optional" json:"serviceAnnotations" yaml:"serviceAnnotations"`
	ServiceIpFamilies *[]interface{} `field:"optional" json:"serviceIpFamilies" yaml:"serviceIpFamilies"`
	ServiceIpFamilyPolicy *string `field:"optional" json:"serviceIpFamilyPolicy" yaml:"serviceIpFamilyPolicy"`
	ServiceLabels interface{} `field:"optional" json:"serviceLabels" yaml:"serviceLabels"`
	ServiceType *string `field:"optional" json:"serviceType" yaml:"serviceType"`
	Strategy interface{} `field:"optional" json:"strategy" yaml:"strategy"`
	TimeoutSeconds *float64 `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints *[]interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	Url interface{} `field:"optional" json:"url" yaml:"url"`
	ValidatingWebhookConfiguration *HelmValuesWebhookValidatingWebhookConfiguration `field:"optional" json:"validatingWebhookConfiguration" yaml:"validatingWebhookConfiguration"`
	ValidatingWebhookConfigurationAnnotations interface{} `field:"optional" json:"validatingWebhookConfigurationAnnotations" yaml:"validatingWebhookConfigurationAnnotations"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

