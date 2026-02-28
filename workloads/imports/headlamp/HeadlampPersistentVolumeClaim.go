package headlamp


type HeadlampPersistentVolumeClaim struct {
	AccessModes *[]*string `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Annotations to add to the persistent volume claim (if enabled).
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	// Enable Persistent Volume Claim.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Selector *HeadlampPersistentVolumeClaimSelector `field:"optional" json:"selector" yaml:"selector"`
	Size *string `field:"optional" json:"size" yaml:"size"`
	StorageClassName *string `field:"optional" json:"storageClassName" yaml:"storageClassName"`
	VolumeMode *string `field:"optional" json:"volumeMode" yaml:"volumeMode"`
}

