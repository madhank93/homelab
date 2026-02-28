package rancher


// The default rancher service configuration.
type RancherService struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	DisableHttp *bool `field:"optional" json:"disableHttp" yaml:"disableHttp"`
	Type RancherServiceType `field:"optional" json:"type" yaml:"type"`
}

