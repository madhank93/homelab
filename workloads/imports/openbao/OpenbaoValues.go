package openbao


type OpenbaoValues struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Csi *OpenbaoCsi `field:"optional" json:"csi" yaml:"csi"`
	Global *map[string]interface{} `field:"optional" json:"global" yaml:"global"`
	Injector *OpenbaoInjector `field:"optional" json:"injector" yaml:"injector"`
	Server *OpenbaoServer `field:"optional" json:"server" yaml:"server"`
	ServerTelemetry *OpenbaoServerTelemetry `field:"optional" json:"serverTelemetry" yaml:"serverTelemetry"`
	Ui *OpenbaoUi `field:"optional" json:"ui" yaml:"ui"`
}

