package grafana


type GrafanaDownloadDashboards struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Env interface{} `field:"optional" json:"env" yaml:"env"`
	EnvFromSecret *string `field:"optional" json:"envFromSecret" yaml:"envFromSecret"`
	EnvValueFrom interface{} `field:"optional" json:"envValueFrom" yaml:"envValueFrom"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
}

