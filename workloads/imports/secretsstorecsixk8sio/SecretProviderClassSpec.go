package secretsstorecsixk8sio


// SecretProviderClassSpec defines the desired state of SecretProviderClass.
type SecretProviderClassSpec struct {
	// Configuration for specific provider.
	Parameters *map[string]*string `field:"optional" json:"parameters" yaml:"parameters"`
	// Configuration for provider name.
	Provider *string `field:"optional" json:"provider" yaml:"provider"`
	SecretObjects *[]*SecretProviderClassSpecSecretObjects `field:"optional" json:"secretObjects" yaml:"secretObjects"`
}

