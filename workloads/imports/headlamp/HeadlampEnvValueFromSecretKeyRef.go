package headlamp


// Selects a key of a Secret.
type HeadlampEnvValueFromSecretKeyRef struct {
	// Key of the Secret to select from.
	Key *string `field:"required" json:"key" yaml:"key"`
	// Name of the Secret.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Specify whether the Secret or its key must be defined.
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

