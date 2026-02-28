package headlamp


// Requests of the init container.
type HeadlampInitContainersResourcesRequests struct {
	// CPU request.
	Cpu *string `field:"optional" json:"cpu" yaml:"cpu"`
	// Memory request.
	Memory *string `field:"optional" json:"memory" yaml:"memory"`
}

