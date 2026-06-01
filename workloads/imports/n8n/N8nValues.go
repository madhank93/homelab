package n8n


type N8nValues struct {
	// DEPRECATED: Use main, worker, and webhook blocks affinity fields instead.
	//
	// This field will be removed in a future release.
	Affinity interface{} `field:"required" json:"affinity" yaml:"affinity"`
	Api *N8NApi `field:"required" json:"api" yaml:"api"`
	// Configuration for binary data storage.
	BinaryData *N8NBinaryData `field:"required" json:"binaryData" yaml:"binaryData"`
	// n8n database configurations.
	Db *N8NDb `field:"required" json:"db" yaml:"db"`
	// A locale identifier, compatible with the Accept-Language header.
	//
	// n8n doesn't support regional identifiers, such as de-AT.
	DefaultLocale N8NDefaultLocale `field:"required" json:"defaultLocale" yaml:"defaultLocale"`
	Diagnostics *N8NDiagnostics `field:"required" json:"diagnostics" yaml:"diagnostics"`
	// For more information checkout: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config.
	DnsConfig *N8NDnsConfig `field:"required" json:"dnsConfig" yaml:"dnsConfig"`
	// For more information checkout: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy.
	DnsPolicy N8NDnsPolicy `field:"required" json:"dnsPolicy" yaml:"dnsPolicy"`
	// If you install n8n first time, you can keep this empty and it will be auto generated and never change again.
	//
	// If you already have a encryption key generated before, please use it here.
	EncryptionKey *string `field:"required" json:"encryptionKey" yaml:"encryptionKey"`
	// The name of an existing secret with encryption key.
	//
	// The secret must contain a key with the name N8N_ENCRYPTION_KEY.
	ExistingEncryptionKeySecret *string `field:"required" json:"existingEncryptionKeySecret" yaml:"existingEncryptionKeySecret"`
	// External PostgreSQL parameters.
	ExternalPostgresql *N8NExternalPostgresql `field:"required" json:"externalPostgresql" yaml:"externalPostgresql"`
	// External Redis parameters.
	ExternalRedis *N8NExternalRedis `field:"required" json:"externalRedis" yaml:"externalRedis"`
	FullnameOverride *string `field:"required" json:"fullnameOverride" yaml:"fullnameOverride"`
	// graceful shutdown timeout in seconds.
	GracefulShutdownTimeout *float64 `field:"required" json:"gracefulShutdownTimeout" yaml:"gracefulShutdownTimeout"`
	// This sets the container image more information can be found here: https://kubernetes.io/docs/concepts/containers/images/.
	Image *N8NImage `field:"required" json:"image" yaml:"image"`
	// This is for the secretes for pulling an image from a private repository more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/.
	ImagePullSecrets *[]interface{} `field:"required" json:"imagePullSecrets" yaml:"imagePullSecrets"`
	// This block is for setting up the ingress for more information can be found here: https://kubernetes.io/docs/concepts/services-networking/ingress/.
	Ingress *N8NIngress `field:"required" json:"ingress" yaml:"ingress"`
	// n8n license configurations.
	License *N8NLicense `field:"required" json:"license" yaml:"license"`
	// n8n log configurations.
	Log *N8NLog `field:"required" json:"log" yaml:"log"`
	// Main node configuration.
	Main *N8NMain `field:"required" json:"main" yaml:"main"`
	Minio *map[string]interface{} `field:"required" json:"minio" yaml:"minio"`
	// This is to override the chart name.
	NameOverride *string `field:"required" json:"nameOverride" yaml:"nameOverride"`
	// Node configurations for built-in and external npm packages.
	Nodes *N8NNodes `field:"required" json:"nodes" yaml:"nodes"`
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector.
	NodeSelector interface{} `field:"required" json:"nodeSelector" yaml:"nodeSelector"`
	// Configuration for private npm registry.
	NpmRegistry *N8NNpmRegistry `field:"required" json:"npmRegistry" yaml:"npmRegistry"`
	// This is for setting Kubernetes Annotations to a Pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
	PodAnnotations interface{} `field:"required" json:"podAnnotations" yaml:"podAnnotations"`
	// This is for setting Kubernetes Labels to a Pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	PodLabels interface{} `field:"required" json:"podLabels" yaml:"podLabels"`
	// This is for setting Security Context to a Pod.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	PodSecurityContext *N8NPodSecurityContext `field:"required" json:"podSecurityContext" yaml:"podSecurityContext"`
	Postgresql *map[string]interface{} `field:"required" json:"postgresql" yaml:"postgresql"`
	Redis *map[string]interface{} `field:"required" json:"redis" yaml:"redis"`
	// This is for setting Security Context to a Container.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	SecurityContext *N8NSecurityContext `field:"required" json:"securityContext" yaml:"securityContext"`
	Sentry *N8NSentry `field:"required" json:"sentry" yaml:"sentry"`
	// This is for setting up a service more information can be found here: https://kubernetes.io/docs/concepts/services-networking/service/.
	Service *N8NService `field:"required" json:"service" yaml:"service"`
	// This section builds out the service account more information can be found here: https://kubernetes.io/docs/concepts/security/service-accounts/.
	ServiceAccount *N8NServiceAccount `field:"required" json:"serviceAccount" yaml:"serviceAccount"`
	// The ServiceMonitor configuration for the n8n deployment.
	//
	// Please refer to the following link for more information: https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md
	ServiceMonitor *N8NServiceMonitor `field:"required" json:"serviceMonitor" yaml:"serviceMonitor"`
	// This will set the deployment strategy more information can be found here: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy.
	Strategy *N8NStrategy `field:"required" json:"strategy" yaml:"strategy"`
	TaskRunners *N8NTaskRunners `field:"required" json:"taskRunners" yaml:"taskRunners"`
	// For instance, the Schedule node uses it to know at what time the workflow should start.
	//
	// Find you timezone from here: https://momentjs.com/timezone/
	Timezone *string `field:"required" json:"timezone" yaml:"timezone"`
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/.
	Tolerations *[]interface{} `field:"required" json:"tolerations" yaml:"tolerations"`
	VersionNotifications *N8NVersionNotifications `field:"required" json:"versionNotifications" yaml:"versionNotifications"`
	Webhook *N8NWebhook `field:"required" json:"webhook" yaml:"webhook"`
	Worker *N8NWorker `field:"required" json:"worker" yaml:"worker"`
	WorkflowHistory *N8NWorkflowHistory `field:"required" json:"workflowHistory" yaml:"workflowHistory"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// DEPRECATED: Use main, worker, and webhook blocks extraEnvVars fields instead.
	//
	// This field will be removed in a future release.
	ExtraEnvVars interface{} `field:"optional" json:"extraEnvVars" yaml:"extraEnvVars"`
	// DEPRECATED: Use main, worker, and webhook blocks extraSecretNamesForEnvFrom fields instead.
	//
	// This field will be removed in a future release.
	ExtraSecretNamesForEnvFrom *[]interface{} `field:"optional" json:"extraSecretNamesForEnvFrom" yaml:"extraSecretNamesForEnvFrom"`
	Global *map[string]interface{} `field:"optional" json:"global" yaml:"global"`
	// DEPRECATED: Use main, worker, and webhook blocks livenessProbe field instead.
	//
	// This field will be removed in a future release.
	LivenessProbe *N8NLivenessProbe `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	// DEPRECATED: Use main, worker, and webhook blocks readinessProbe field instead.
	//
	// This field will be removed in a future release.
	ReadinessProbe *N8NReadinessProbe `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	// DEPRECATED: Use main, worker, and webhook blocks resources fields instead.
	//
	// This field will be removed in a future release.
	Resources *N8NResources `field:"optional" json:"resources" yaml:"resources"`
	// DEPRECATED: Use main, worker, and webhook blocks volumeMounts fields instead.
	//
	// This field will be removed in a future release.
	VolumeMounts *[]interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	// DEPRECATED: Use main, worker, and webhook blocks volumes fields instead.
	//
	// This field will be removed in a future release.
	Volumes *[]interface{} `field:"optional" json:"volumes" yaml:"volumes"`
}

