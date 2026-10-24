package grafana


type GrafanaImageRendererService struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	AppProtocol *string `field:"optional" json:"appProtocol" yaml:"appProtocol"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	PortName *string `field:"optional" json:"portName" yaml:"portName"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

