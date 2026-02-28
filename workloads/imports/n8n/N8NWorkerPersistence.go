package n8n


// Persistence configuration for the worker pod.
type N8NWorkerPersistence struct {
	// Access mode for persistence.
	AccessMode N8NWorkerPersistenceAccessMode `field:"required" json:"accessMode" yaml:"accessMode"`
	// Annotations for persistence.
	Annotations interface{} `field:"required" json:"annotations" yaml:"annotations"`
	// Whether to enable persistence.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Existing claim to use for persistence.
	ExistingClaim *string `field:"required" json:"existingClaim" yaml:"existingClaim"`
	// Labels for persistence.
	Labels interface{} `field:"required" json:"labels" yaml:"labels"`
	// Mount path for persistence.
	MountPath *string `field:"required" json:"mountPath" yaml:"mountPath"`
	// Size for persistence.
	Size *string `field:"required" json:"size" yaml:"size"`
	// Storage class for persistence.
	StorageClass *string `field:"required" json:"storageClass" yaml:"storageClass"`
	// Sub path for persistence.
	SubPath *string `field:"required" json:"subPath" yaml:"subPath"`
	// Name of the volume to use for persistence.
	VolumeName *string `field:"required" json:"volumeName" yaml:"volumeName"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

