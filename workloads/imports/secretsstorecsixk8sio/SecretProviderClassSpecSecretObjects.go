package secretsstorecsixk8sio


// SecretObject defines the desired state of synced K8s secret objects.
type SecretProviderClassSpecSecretObjects struct {
	// annotations of k8s secret object.
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	Data *[]*SecretProviderClassSpecSecretObjectsData `field:"optional" json:"data" yaml:"data"`
	// labels of K8s secret object.
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// name of the K8s secret object.
	SecretName *string `field:"optional" json:"secretName" yaml:"secretName"`
	// type of K8s secret object.
	Type *string `field:"optional" json:"type" yaml:"type"`
}

