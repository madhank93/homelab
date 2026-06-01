package rancher


// The default rancher network exposure configuration.
type RancherNetworkExposure struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Type RancherNetworkExposureType `field:"optional" json:"type" yaml:"type"`
}

