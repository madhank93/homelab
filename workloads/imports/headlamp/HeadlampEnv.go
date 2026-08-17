package headlamp


type HeadlampEnv struct {
	// Name of the environment variable.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Value of the environment variable.
	Value *string `field:"optional" json:"value" yaml:"value"`
	// Source for the environment variable's value.
	ValueFrom *HeadlampEnvValueFrom `field:"optional" json:"valueFrom" yaml:"valueFrom"`
}

