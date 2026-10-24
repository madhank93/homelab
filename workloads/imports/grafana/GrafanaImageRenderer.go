package grafana


type GrafanaImageRenderer struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Autoscaling *GrafanaImageRendererAutoscaling `field:"optional" json:"autoscaling" yaml:"autoscaling"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	DeploymentStrategy interface{} `field:"optional" json:"deploymentStrategy" yaml:"deploymentStrategy"`
	DnsConfig interface{} `field:"optional" json:"dnsConfig" yaml:"dnsConfig"`
	DnsPolicy *string `field:"optional" json:"dnsPolicy" yaml:"dnsPolicy"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Env interface{} `field:"optional" json:"env" yaml:"env"`
	EnvValueFrom interface{} `field:"optional" json:"envValueFrom" yaml:"envValueFrom"`
	ExistingSecret *string `field:"optional" json:"existingSecret" yaml:"existingSecret"`
	ExtraConfigmapMounts *[]interface{} `field:"optional" json:"extraConfigmapMounts" yaml:"extraConfigmapMounts"`
	ExtraSecretMounts *[]interface{} `field:"optional" json:"extraSecretMounts" yaml:"extraSecretMounts"`
	ExtraVolumeMounts *[]interface{} `field:"optional" json:"extraVolumeMounts" yaml:"extraVolumeMounts"`
	ExtraVolumes *[]interface{} `field:"optional" json:"extraVolumes" yaml:"extraVolumes"`
	GrafanaProtocol *string `field:"optional" json:"grafanaProtocol" yaml:"grafanaProtocol"`
	GrafanaSubPath *string `field:"optional" json:"grafanaSubPath" yaml:"grafanaSubPath"`
	HealthcheckPath *string `field:"optional" json:"healthcheckPath" yaml:"healthcheckPath"`
	HostAliases *[]interface{} `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostUsers *bool `field:"optional" json:"hostUsers" yaml:"hostUsers"`
	Image *GrafanaImageRendererImage `field:"optional" json:"image" yaml:"image"`
	NetworkPolicy *GrafanaImageRendererNetworkPolicy `field:"optional" json:"networkPolicy" yaml:"networkPolicy"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PodAnnotations *map[string]*string `field:"optional" json:"podAnnotations" yaml:"podAnnotations"`
	PodPortName *string `field:"optional" json:"podPortName" yaml:"podPortName"`
	PriorityClassName *string `field:"optional" json:"priorityClassName" yaml:"priorityClassName"`
	ProgressDeadlineSeconds *float64 `field:"optional" json:"progressDeadlineSeconds" yaml:"progressDeadlineSeconds"`
	ReadinessProbe interface{} `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	RenderingCallbackUrl *string `field:"optional" json:"renderingCallbackUrl" yaml:"renderingCallbackUrl"`
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	RevisionHistoryLimit *float64 `field:"optional" json:"revisionHistoryLimit" yaml:"revisionHistoryLimit"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServerUrl *string `field:"optional" json:"serverUrl" yaml:"serverUrl"`
	Service *GrafanaImageRendererService `field:"optional" json:"service" yaml:"service"`
	ServiceAccountName *string `field:"optional" json:"serviceAccountName" yaml:"serviceAccountName"`
	ServiceMonitor *GrafanaImageRendererServiceMonitor `field:"optional" json:"serviceMonitor" yaml:"serviceMonitor"`
	Token *string `field:"optional" json:"token" yaml:"token"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
}

