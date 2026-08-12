package headlamp


type HeadlampConfigClusterInventoryPlugins struct {
	// Image reference for the plugin image volume.
	Image *string `field:"required" json:"image" yaml:"image"`
	// Absolute read-only mount path for the plugin image volume.
	MountPath *string `field:"required" json:"mountPath" yaml:"mountPath"`
	// DNS label name of the plugin image volume.
	Name *string `field:"required" json:"name" yaml:"name"`
}

