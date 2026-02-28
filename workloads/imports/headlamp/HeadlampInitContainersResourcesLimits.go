package headlamp


// Limits of the init container.
type HeadlampInitContainersResourcesLimits struct {
	// CPU limit.
	Cpu *string `field:"optional" json:"cpu" yaml:"cpu"`
	// Memory limit.
	Memory *string `field:"optional" json:"memory" yaml:"memory"`
}

