package headlamp


type HeadlampValues struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Mount Service Account token in pod.
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	ClusterRoleBinding *HeadlampClusterRoleBinding `field:"optional" json:"clusterRoleBinding" yaml:"clusterRoleBinding"`
	// Headlamp deployment configuration.
	Config *HeadlampConfig `field:"optional" json:"config" yaml:"config"`
	// Environment variables to pass to the deployment.
	Env *[]*HeadlampEnv `field:"optional" json:"env" yaml:"env"`
	// Extra manifests to apply to the deployment.
	ExtraManifests *[]*string `field:"optional" json:"extraManifests" yaml:"extraManifests"`
	// Override the full name of the chart.
	FullnameOverride *string `field:"optional" json:"fullnameOverride" yaml:"fullnameOverride"`
	Global *map[string]interface{} `field:"optional" json:"global" yaml:"global"`
	// HTTPRoute configuration for Gateway API.
	HttpRoute *HeadlampHttpRoute `field:"optional" json:"httpRoute" yaml:"httpRoute"`
	// Image to deploy.
	Image *HeadlampImage `field:"optional" json:"image" yaml:"image"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	//
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	ImagePullSecrets *[]*HeadlampImagePullSecrets `field:"optional" json:"imagePullSecrets" yaml:"imagePullSecrets"`
	Ingress *HeadlampIngress `field:"optional" json:"ingress" yaml:"ingress"`
	// Init containers.
	InitContainers *[]*HeadlampInitContainers `field:"optional" json:"initContainers" yaml:"initContainers"`
	// Override the name of the chart.
	NameOverride *string `field:"optional" json:"nameOverride" yaml:"nameOverride"`
	// Override the deployment namespace;
	//
	// defaults to .Release.Namespace
	NamespaceOverride *string `field:"optional" json:"namespaceOverride" yaml:"namespaceOverride"`
	PersistentVolumeClaim *HeadlampPersistentVolumeClaim `field:"optional" json:"persistentVolumeClaim" yaml:"persistentVolumeClaim"`
	PodDisruptionBudget *HeadlampPodDisruptionBudget `field:"optional" json:"podDisruptionBudget" yaml:"podDisruptionBudget"`
	// Number of replicas to deploy.
	ReplicaCount *float64 `field:"optional" json:"replicaCount" yaml:"replicaCount"`
	Service *HeadlampService `field:"optional" json:"service" yaml:"service"`
	ServiceAccount *HeadlampServiceAccount `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
}

