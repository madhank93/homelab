package openbao


type OpenbaoServerDataStorage struct {
	AccessMode *string `field:"optional" json:"accessMode" yaml:"accessMode"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	MountPath *string `field:"optional" json:"mountPath" yaml:"mountPath"`
	Size *string `field:"optional" json:"size" yaml:"size"`
	StorageClass *string `field:"optional" json:"storageClass" yaml:"storageClass"`
}

