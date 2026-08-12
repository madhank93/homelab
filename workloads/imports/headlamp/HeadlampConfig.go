package headlamp


// Headlamp deployment configuration.
type HeadlampConfig struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Base URL of the application.
	BaseUrl *string `field:"optional" json:"baseUrl" yaml:"baseUrl"`
	// Experimental/alpha Cluster Inventory configuration.
	ClusterInventory *HeadlampConfigClusterInventory `field:"optional" json:"clusterInventory" yaml:"clusterInventory"`
	// Extra arguments to pass to the application.
	ExtraArgs *[]*string `field:"optional" json:"extraArgs" yaml:"extraArgs"`
	// Default image to use when creating node shell pods.
	NodeShellImage *string `field:"optional" json:"nodeShellImage" yaml:"nodeShellImage"`
	// Default namespace to use when creating node shell pods.
	NodeShellNamespace *string `field:"optional" json:"nodeShellNamespace" yaml:"nodeShellNamespace"`
	// OIDC configuration.
	Oidc *HeadlampConfigOidc `field:"optional" json:"oidc" yaml:"oidc"`
	// Directory to load plugins from.
	PluginsDir *string `field:"optional" json:"pluginsDir" yaml:"pluginsDir"`
	// Default image to use when creating pod debug containers.
	PodDebugImage *string `field:"optional" json:"podDebugImage" yaml:"podDebugImage"`
	// Path to the service account token file when unsafeUseServiceAccountToken is enabled.
	ServiceAccountTokenPath *string `field:"optional" json:"serviceAccountTokenPath" yaml:"serviceAccountTokenPath"`
	// The time in seconds for the session to be valid.
	SessionTtl *float64 `field:"optional" json:"sessionTtl" yaml:"sessionTtl"`
	// Bundled (static) plugins shipped in the Headlamp image, such as the Prometheus plugin.
	StaticPlugins *HeadlampConfigStaticPlugins `field:"optional" json:"staticPlugins" yaml:"staticPlugins"`
	// Path of certificate file for TLS.
	TlsCertPath *string `field:"optional" json:"tlsCertPath" yaml:"tlsCertPath"`
	// Path of private key file for TLS.
	TlsKeyPath *string `field:"optional" json:"tlsKeyPath" yaml:"tlsKeyPath"`
	// UNSAFE: authenticate every user as the pod's service account in-cluster mode.
	UnsafeUseServiceAccountToken *bool `field:"optional" json:"unsafeUseServiceAccountToken" yaml:"unsafeUseServiceAccountToken"`
}

