package grafana


type GrafanaPersistence struct {
	AccessModes *[]*string `field:"optional" json:"accessModes" yaml:"accessModes"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	DisableWarning *bool `field:"optional" json:"disableWarning" yaml:"disableWarning"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraPvcLabels *map[string]*string `field:"optional" json:"extraPvcLabels" yaml:"extraPvcLabels"`
	Finalizers *[]*string `field:"optional" json:"finalizers" yaml:"finalizers"`
	InMemory *GrafanaPersistenceInMemory `field:"optional" json:"inMemory" yaml:"inMemory"`
	LookupVolumeName *bool `field:"optional" json:"lookupVolumeName" yaml:"lookupVolumeName"`
	PersistentVolumeClaimRetentionPolicy interface{} `field:"optional" json:"persistentVolumeClaimRetentionPolicy" yaml:"persistentVolumeClaimRetentionPolicy"`
	Size interface{} `field:"optional" json:"size" yaml:"size"`
	Type *string `field:"optional" json:"type" yaml:"type"`
	VolumeName *string `field:"optional" json:"volumeName" yaml:"volumeName"`
}

