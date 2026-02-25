package headlamp


type HeadlampInitContainersVolumeMounts struct {
	// Mount path of the volume mount.
	MountPath *string `field:"optional" json:"mountPath" yaml:"mountPath"`
	// Name of the volume mount.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// Read only of the volume mount.
	ReadOnly *bool `field:"optional" json:"readOnly" yaml:"readOnly"`
}

