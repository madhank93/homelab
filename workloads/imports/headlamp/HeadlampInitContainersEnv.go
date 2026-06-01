package headlamp


type HeadlampInitContainersEnv struct {
	// Name of the environment variable.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Value of the environment variable.
	Value *string `field:"optional" json:"value" yaml:"value"`
}

