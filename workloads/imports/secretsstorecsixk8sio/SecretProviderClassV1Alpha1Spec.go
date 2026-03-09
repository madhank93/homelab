package secretsstorecsixk8sio


// SecretProviderClassSpec defines the desired state of SecretProviderClass.
type SecretProviderClassV1Alpha1Spec struct {
	// Configuration for specific provider.
	Parameters *map[string]*string `field:"optional" json:"parameters" yaml:"parameters"`
	// Configuration for provider name.
	Provider *string `field:"optional" json:"provider" yaml:"provider"`
	SecretObjects *[]*SecretProviderClassV1Alpha1SpecSecretObjects `field:"optional" json:"secretObjects" yaml:"secretObjects"`
}

