package openbao


type OpenbaoServerPersistentVolumeClaimRetentionPolicy struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	WhenDeleted *string `field:"optional" json:"whenDeleted" yaml:"whenDeleted"`
	WhenScaled *string `field:"optional" json:"whenScaled" yaml:"whenScaled"`
}

