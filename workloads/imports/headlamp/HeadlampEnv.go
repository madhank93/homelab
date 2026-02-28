package headlamp


type HeadlampEnv struct {
	// Name of the environment variable.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Value of the environment variable.
	Value *string `field:"required" json:"value" yaml:"value"`
}

