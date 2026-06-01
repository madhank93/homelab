package openbao


type OpenbaoServer struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	AuditStorage *OpenbaoServerAuditStorage `field:"optional" json:"auditStorage" yaml:"auditStorage"`
	AuthDelegator *OpenbaoServerAuthDelegator `field:"optional" json:"authDelegator" yaml:"authDelegator"`
	DataStorage *OpenbaoServerDataStorage `field:"optional" json:"dataStorage" yaml:"dataStorage"`
	Dev *OpenbaoServerDev `field:"optional" json:"dev" yaml:"dev"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraArgs *string `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	ExtraContainers *[]interface{} `field:"optional" json:"extraContainers" yaml:"extraContainers"`
	ExtraEnvironmentVars interface{} `field:"optional" json:"extraEnvironmentVars" yaml:"extraEnvironmentVars"`
	ExtraInitContainers *[]interface{} `field:"optional" json:"extraInitContainers" yaml:"extraInitContainers"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	ExtraPorts *[]interface{} `field:"optional" json:"extraPorts" yaml:"extraPorts"`
	ExtraSecretEnvironmentVars *[]interface{} `field:"optional" json:"extraSecretEnvironmentVars" yaml:"extraSecretEnvironmentVars"`
	ExtraVolumes *[]interface{} `field:"optional" json:"extraVolumes" yaml:"extraVolumes"`
	Ha *OpenbaoServerHa `field:"optional" json:"ha" yaml:"ha"`
	HostAliases *[]interface{} `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostNetwork *bool `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	Image *OpenbaoServerImage `field:"optional" json:"image" yaml:"image"`
	Ingress *OpenbaoServerIngress `field:"optional" json:"ingress" yaml:"ingress"`
	LivenessProbe *OpenbaoServerLivenessProbe `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	LogFormat *string `field:"optional" json:"logFormat" yaml:"logFormat"`
	LogLevel *string `field:"optional" json:"logLevel" yaml:"logLevel"`
	NetworkPolicy *OpenbaoServerNetworkPolicy `field:"optional" json:"networkPolicy" yaml:"networkPolicy"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PersistentVolumeClaimRetentionPolicy *OpenbaoServerPersistentVolumeClaimRetentionPolicy `field:"optional" json:"persistentVolumeClaimRetentionPolicy" yaml:"persistentVolumeClaimRetentionPolicy"`
	PostStart *[]interface{} `field:"optional" json:"postStart" yaml:"postStart"`
	PreStopSleepSeconds *float64 `field:"optional" json:"preStopSleepSeconds" yaml:"preStopSleepSeconds"`
	PriorityClassName *string `field:"optional" json:"priorityClassName" yaml:"priorityClassName"`
	ReadinessProbe *OpenbaoServerReadinessProbe `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	Route *OpenbaoServerRoute `field:"optional" json:"route" yaml:"route"`
	Service *OpenbaoServerService `field:"optional" json:"service" yaml:"service"`
	ServiceAccount *OpenbaoServerServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace *bool `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Standalone *OpenbaoServerStandalone `field:"optional" json:"standalone" yaml:"standalone"`
	StatefulSet *OpenbaoServerStatefulSet `field:"optional" json:"statefulSet" yaml:"statefulSet"`
	TerminationGracePeriodSeconds *float64 `field:"optional" json:"terminationGracePeriodSeconds" yaml:"terminationGracePeriodSeconds"`
	Tolerations interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	UpdateStrategyType *string `field:"optional" json:"updateStrategyType" yaml:"updateStrategyType"`
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

