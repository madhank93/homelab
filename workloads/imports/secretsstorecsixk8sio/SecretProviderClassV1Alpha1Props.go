package secretsstorecsixk8sio

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

// SecretProviderClass is the Schema for the secretproviderclasses API.
type SecretProviderClassV1Alpha1Props struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// SecretProviderClassSpec defines the desired state of SecretProviderClass.
	Spec *SecretProviderClassV1Alpha1Spec `field:"optional" json:"spec" yaml:"spec"`
}

