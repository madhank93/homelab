package headlamp


type HeadlampInitContainers struct {
	// Arguments of the init container.
	Args *[]*string `field:"optional" json:"args" yaml:"args"`
	// Command of the init container.
	Command *[]*string `field:"optional" json:"command" yaml:"command"`
	// Environment variables of the init container.
	Env *[]*HeadlampInitContainersEnv `field:"optional" json:"env" yaml:"env"`
	// Image of the init container.
	Image *string `field:"optional" json:"image" yaml:"image"`
	// Pull policy of the init container.
	ImagePullPolicy HeadlampInitContainersImagePullPolicy `field:"optional" json:"imagePullPolicy" yaml:"imagePullPolicy"`
	// Name of the init container.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Resources of the init container.
	Resources *HeadlampInitContainersResources `field:"optional" json:"resources" yaml:"resources"`
	// Volume mounts of the init container.
	VolumeMounts *[]*HeadlampInitContainersVolumeMounts `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
}

