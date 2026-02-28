package certmanager


type HelmValues struct {
	Acmesolver *HelmValuesAcmesolver `field:"optional" json:"acmesolver" yaml:"acmesolver"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	ApproveSignerNames *[]interface{} `field:"optional" json:"approveSignerNames" yaml:"approveSignerNames"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Cainjector *HelmValuesCainjector `field:"optional" json:"cainjector" yaml:"cainjector"`
	ClusterResourceNamespace *string `field:"optional" json:"clusterResourceNamespace" yaml:"clusterResourceNamespace"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	Crds *HelmValuesCrds `field:"optional" json:"crds" yaml:"crds"`
	Creator *string `field:"optional" json:"creator" yaml:"creator"`
	DeploymentAnnotations interface{} `field:"optional" json:"deploymentAnnotations" yaml:"deploymentAnnotations"`
	DisableAutoApproval *bool `field:"optional" json:"disableAutoApproval" yaml:"disableAutoApproval"`
	Dns01RecursiveNameservers *string `field:"optional" json:"dns01RecursiveNameservers" yaml:"dns01RecursiveNameservers"`
	Dns01RecursiveNameserversOnly *bool `field:"optional" json:"dns01RecursiveNameserversOnly" yaml:"dns01RecursiveNameserversOnly"`
	EnableCertificateOwnerRef *bool `field:"optional" json:"enableCertificateOwnerRef" yaml:"enableCertificateOwnerRef"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	ExtraArgs *[]interface{} `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	ExtraEnv *[]interface{} `field:"optional" json:"extraEnv" yaml:"extraEnv"`
	ExtraObjects *[]interface{} `field:"optional" json:"extraObjects" yaml:"extraObjects"`
	FeatureGates *string `field:"optional" json:"featureGates" yaml:"featureGates"`
	FullnameOverride *string `field:"optional" json:"fullnameOverride" yaml:"fullnameOverride"`
	Global *HelmValuesGlobal `field:"optional" json:"global" yaml:"global"`
	HostAliases *[]interface{} `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HttpProxy *string `field:"optional" json:"httpProxy" yaml:"httpProxy"`
	HttpsProxy *string `field:"optional" json:"httpsProxy" yaml:"httpsProxy"`
	Image *HelmValuesImage `field:"optional" json:"image" yaml:"image"`
	IngressShim *HelmValuesIngressShim `field:"optional" json:"ingressShim" yaml:"ingressShim"`
	InstallCrDs *bool `field:"optional" json:"installCrDs" yaml:"installCrDs"`
	LivenessProbe interface{} `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	MaxConcurrentChallenges *float64 `field:"optional" json:"maxConcurrentChallenges" yaml:"maxConcurrentChallenges"`
	NameOverride *string `field:"optional" json:"nameOverride" yaml:"nameOverride"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	NoProxy *string `field:"optional" json:"noProxy" yaml:"noProxy"`
	PodAnnotations interface{} `field:"optional" json:"podAnnotations" yaml:"podAnnotations"`
	PodDisruptionBudget *HelmValuesPodDisruptionBudget `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	PodDnsConfig interface{} `field:"optional" json:"podDnsConfig" yaml:"podDnsConfig"`
	PodDnsPolicy *string `field:"optional" json:"podDnsPolicy" yaml:"podDnsPolicy"`
	PodLabels interface{} `field:"optional" json:"podLabels" yaml:"podLabels"`
	Prometheus *HelmValuesPrometheus `field:"optional" json:"prometheus" yaml:"prometheus"`
	ReplicaCount *float64 `field:"optional" json:"replicaCount" yaml:"replicaCount"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount *HelmValuesServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ServiceAnnotations interface{} `field:"optional" json:"serviceAnnotations" yaml:"serviceAnnotations"`
	ServiceIpFamilies *[]interface{} `field:"optional" json:"serviceIpFamilies" yaml:"serviceIpFamilies"`
	ServiceIpFamilyPolicy *string `field:"optional" json:"serviceIpFamilyPolicy" yaml:"serviceIpFamilyPolicy"`
	ServiceLabels interface{} `field:"optional" json:"serviceLabels" yaml:"serviceLabels"`
	Startupapicheck *HelmValuesStartupapicheck `field:"optional" json:"startupapicheck" yaml:"startupapicheck"`
	Strategy interface{} `field:"optional" json:"strategy" yaml:"strategy"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints *[]interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
	Webhook *HelmValuesWebhook `field:"optional" json:"webhook" yaml:"webhook"`
}

