package headlamp


// Selects a resource of the container.
type HeadlampEnvValueFromResourceFieldRef struct {
	// Resource to select.
	Resource *string `field:"required" json:"resource" yaml:"resource"`
	// Container name to select resources from.
	ContainerName *string `field:"optional" json:"containerName" yaml:"containerName"`
	// Output format of the exposed resource.
	Divisor *string `field:"optional" json:"divisor" yaml:"divisor"`
}

