package grafana


type GrafanaValues struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Admin *GrafanaAdmin `field:"optional" json:"admin" yaml:"admin"`
	AdminUser *string `field:"optional" json:"adminUser" yaml:"adminUser"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	Alerting interface{} `field:"optional" json:"alerting" yaml:"alerting"`
	AssertNoLeakedSecrets *bool `field:"optional" json:"assertNoLeakedSecrets" yaml:"assertNoLeakedSecrets"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Autoscaling *GrafanaAutoscaling `field:"optional" json:"autoscaling" yaml:"autoscaling"`
	ContainerSecurityContext interface{} `field:"optional" json:"containerSecurityContext" yaml:"containerSecurityContext"`
	CreateConfigmap *bool `field:"optional" json:"createConfigmap" yaml:"createConfigmap"`
	DashboardProviders interface{} `field:"optional" json:"dashboardProviders" yaml:"dashboardProviders"`
	Dashboards interface{} `field:"optional" json:"dashboards" yaml:"dashboards"`
	DashboardsConfigMaps interface{} `field:"optional" json:"dashboardsConfigMaps" yaml:"dashboardsConfigMaps"`
	Datasources interface{} `field:"optional" json:"datasources" yaml:"datasources"`
	DefaultCurlOptions *string `field:"optional" json:"defaultCurlOptions" yaml:"defaultCurlOptions"`
	DefaultShellOptions *string `field:"optional" json:"defaultShellOptions" yaml:"defaultShellOptions"`
	DeploymentStrategy interface{} `field:"optional" json:"deploymentStrategy" yaml:"deploymentStrategy"`
	DnsConfig interface{} `field:"optional" json:"dnsConfig" yaml:"dnsConfig"`
	DnsPolicy *string `field:"optional" json:"dnsPolicy" yaml:"dnsPolicy"`
	DownloadDashboards *GrafanaDownloadDashboards `field:"optional" json:"downloadDashboards" yaml:"downloadDashboards"`
	DownloadDashboardsImage *GrafanaDownloadDashboardsImage `field:"optional" json:"downloadDashboardsImage" yaml:"downloadDashboardsImage"`
	EnableKubeBackwardCompatibility *bool `field:"optional" json:"enableKubeBackwardCompatibility" yaml:"enableKubeBackwardCompatibility"`
	EnableServiceLinks *bool `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	Env interface{} `field:"optional" json:"env" yaml:"env"`
	EnvFromConfigMaps *[]interface{} `field:"optional" json:"envFromConfigMaps" yaml:"envFromConfigMaps"`
	EnvFromSecret *string `field:"optional" json:"envFromSecret" yaml:"envFromSecret"`
	EnvFromSecrets *[]interface{} `field:"optional" json:"envFromSecrets" yaml:"envFromSecrets"`
	EnvRenderSecret interface{} `field:"optional" json:"envRenderSecret" yaml:"envRenderSecret"`
	EnvValueFrom interface{} `field:"optional" json:"envValueFrom" yaml:"envValueFrom"`
	ExtraConfigmapMounts *[]interface{} `field:"optional" json:"extraConfigmapMounts" yaml:"extraConfigmapMounts"`
	ExtraContainers interface{} `field:"optional" json:"extraContainers" yaml:"extraContainers"`
	ExtraContainerVolumes *[]interface{} `field:"optional" json:"extraContainerVolumes" yaml:"extraContainerVolumes"`
	ExtraEmptyDirMounts *[]interface{} `field:"optional" json:"extraEmptyDirMounts" yaml:"extraEmptyDirMounts"`
	ExtraExposePorts *[]interface{} `field:"optional" json:"extraExposePorts" yaml:"extraExposePorts"`
	ExtraInitContainers *[]interface{} `field:"optional" json:"extraInitContainers" yaml:"extraInitContainers"`
	ExtraLabels *map[string]*string `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	ExtraObjects *[]interface{} `field:"optional" json:"extraObjects" yaml:"extraObjects"`
	ExtraSecretMounts *[]interface{} `field:"optional" json:"extraSecretMounts" yaml:"extraSecretMounts"`
	ExtraVolumeMounts *[]interface{} `field:"optional" json:"extraVolumeMounts" yaml:"extraVolumeMounts"`
	ExtraVolumes *[]interface{} `field:"optional" json:"extraVolumes" yaml:"extraVolumes"`
	Global *map[string]interface{} `field:"optional" json:"global" yaml:"global"`
	GoMemLimit *GrafanaGoMemLimit `field:"optional" json:"goMemLimit" yaml:"goMemLimit"`
	GossipPortName *string `field:"optional" json:"gossipPortName" yaml:"gossipPortName"`
	GrafanaIni interface{} `field:"optional" json:"grafanaIni" yaml:"grafanaIni"`
	HeadlessService *bool `field:"optional" json:"headlessService" yaml:"headlessService"`
	HostAliases *[]interface{} `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostUsers *bool `field:"optional" json:"hostUsers" yaml:"hostUsers"`
	Image *GrafanaImage `field:"optional" json:"image" yaml:"image"`
	ImageRenderer *GrafanaImageRenderer `field:"optional" json:"imageRenderer" yaml:"imageRenderer"`
	Ingress *GrafanaIngress `field:"optional" json:"ingress" yaml:"ingress"`
	InitChownData *GrafanaInitChownData `field:"optional" json:"initChownData" yaml:"initChownData"`
	Ldap *GrafanaLdap `field:"optional" json:"ldap" yaml:"ldap"`
	LifecycleHooks interface{} `field:"optional" json:"lifecycleHooks" yaml:"lifecycleHooks"`
	// BETA: Configure Gateway API ListenerSet resources for the chart here.
	//
	// ListenerSet allows attaching additional listeners to an existing Gateway. More listener sets can be added by adding a dictionary key like the 'main' entry. Being BETA this can/will change in the future without notice, do not use unless you want to take that risk [[ref]](https://gateway-api.sigs.k8s.io/reference/api-spec/main/spec/#listenerset)
	ListenerSet *GrafanaListenerSet `field:"optional" json:"listenerSet" yaml:"listenerSet"`
	LivenessProbe interface{} `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	NamespaceOverride *string `field:"optional" json:"namespaceOverride" yaml:"namespaceOverride"`
	NetworkPolicy *GrafanaNetworkPolicy `field:"optional" json:"networkPolicy" yaml:"networkPolicy"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	Notifiers interface{} `field:"optional" json:"notifiers" yaml:"notifiers"`
	Persistence *GrafanaPersistence `field:"optional" json:"persistence" yaml:"persistence"`
	Plugins *[]interface{} `field:"optional" json:"plugins" yaml:"plugins"`
	PodDisruptionBudget interface{} `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	PodManagementPolicy *string `field:"optional" json:"podManagementPolicy" yaml:"podManagementPolicy"`
	PodPortName *string `field:"optional" json:"podPortName" yaml:"podPortName"`
	ProgressDeadlineSeconds *float64 `field:"optional" json:"progressDeadlineSeconds" yaml:"progressDeadlineSeconds"`
	Rbac *GrafanaRbac `field:"optional" json:"rbac" yaml:"rbac"`
	ReadinessProbe interface{} `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	RevisionHistoryLimit *float64 `field:"optional" json:"revisionHistoryLimit" yaml:"revisionHistoryLimit"`
	// BETA: Configure the gateway routes for the chart here.
	//
	// More routes can be added by adding a dictionary key like the 'main' route. Be aware that this is an early beta of this feature, kube-prometheus-stack does not guarantee this works and is subject to change. Being BETA this can/will change in the future without notice, do not use unless you want to take that risk [[ref]](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io%2fv1alpha2)
	Route *GrafanaRoute `field:"optional" json:"route" yaml:"route"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
	Service *GrafanaService `field:"optional" json:"service" yaml:"service"`
	ServiceAccount *GrafanaServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ServiceMonitor *GrafanaServiceMonitor `field:"optional" json:"serviceMonitor" yaml:"serviceMonitor"`
	ShadowBundledPlugins *bool `field:"optional" json:"shadowBundledPlugins" yaml:"shadowBundledPlugins"`
	ShareProcessNamespace *bool `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	Sidecar *GrafanaSidecar `field:"optional" json:"sidecar" yaml:"sidecar"`
	Smtp *GrafanaSmtp `field:"optional" json:"smtp" yaml:"smtp"`
	StartupProbe interface{} `field:"optional" json:"startupProbe" yaml:"startupProbe"`
	TestFramework *GrafanaTestFramework `field:"optional" json:"testFramework" yaml:"testFramework"`
	Tolerations *[]interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
	TopologySpreadConstraints *[]interface{} `field:"optional" json:"topologySpreadConstraints" yaml:"topologySpreadConstraints"`
	UseStatefulSet *bool `field:"optional" json:"useStatefulSet" yaml:"useStatefulSet"`
	VerticalPodAutoscaler *GrafanaVerticalPodAutoscaler `field:"optional" json:"verticalPodAutoscaler" yaml:"verticalPodAutoscaler"`
}

