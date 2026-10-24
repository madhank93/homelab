package grafana


type GrafanaImageRendererNetworkPolicy struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ExtraIngressSelectors *[]interface{} `field:"optional" json:"extraIngressSelectors" yaml:"extraIngressSelectors"`
	LimitEgress *bool `field:"optional" json:"limitEgress" yaml:"limitEgress"`
	LimitIngress *bool `field:"optional" json:"limitIngress" yaml:"limitIngress"`
}

