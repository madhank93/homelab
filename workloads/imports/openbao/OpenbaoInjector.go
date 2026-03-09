package openbao


type OpenbaoInjector struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	AgentDefaults *OpenbaoInjectorAgentDefaults `field:"optional" json:"agentDefaults" yaml:"agentDefaults"`
	AgentImage *OpenbaoInjectorAgentImage `field:"optional" json:"agentImage" yaml:"agentImage"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	AuthPath *string `field:"optional" json:"authPath" yaml:"authPath"`
	Certs *OpenbaoInjectorCerts `field:"optional" json:"certs" yaml:"certs"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	ExternalVaultAddr *string `field:"optional" json:"externalVaultAddr" yaml:"externalVaultAddr"`
	ExtraEnvironmentVars interface{} `field:"optional" json:"extraEnvironmentVars" yaml:"extraEnvironmentVars"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	FailurePolicy *string `field:"optional" json:"failurePolicy" yaml:"failurePolicy"`
	HostNetwork *bool `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	Image *OpenbaoInjectorImage `field:"optional" json:"image" yaml:"image"`
	LeaderElector *OpenbaoInjectorLeaderElector `field:"optional" json:"leaderElector" yaml:"leaderElector"`
	LogFormat *string `field:"optional" json:"logFormat" yaml:"logFormat"`
	LogLevel *string `field:"optional" json:"logLevel" yaml:"logLevel"`
	Metrics *OpenbaoInjectorMetrics `field:"optional" json:"metrics" yaml:"metrics"`
	NamespaceSelector interface{} `field:"optional" json:"namespaceSelector" yaml:"namespaceSelector"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	ObjectSelector interface{} `field:"optional" json:"objectSelector" yaml:"objectSelector"`
	PodDisruptionBudget interface{} `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	PriorityClassName *string `field:"optional" json:"priorityClassName" yaml:"priorityClassName"`
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	RevokeOnShutdown *bool `field:"optional" json:"revokeOnShutdown" yaml:"revokeOnShutdown"`
	SecurityContext *OpenbaoInjectorSecurityContext `field:"optional" json:"securityContext" yaml:"securityContext"`
	Service *OpenbaoInjectorService `field:"optional" json:"service" yaml:"service"`
	ServiceAccount *OpenbaoInjectorServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	Strategy interface{} `field:"optional" json:"strategy" yaml:"strategy"`
	Tolerations interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	Webhook *OpenbaoInjectorWebhook `field:"optional" json:"webhook" yaml:"webhook"`
	WebhookAnnotations interface{} `field:"optional" json:"webhookAnnotations" yaml:"webhookAnnotations"`
}

