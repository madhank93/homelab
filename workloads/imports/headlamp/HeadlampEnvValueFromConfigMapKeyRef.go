package headlamp


// Selects a key of a ConfigMap.
type HeadlampEnvValueFromConfigMapKeyRef struct {
	// Key of the ConfigMap to select from.
	Key *string `field:"required" json:"key" yaml:"key"`
	// Name of the ConfigMap.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Specify whether the ConfigMap or its key must be defined.
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

