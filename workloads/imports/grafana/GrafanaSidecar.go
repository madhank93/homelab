package grafana


type GrafanaSidecar struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Alerts *GrafanaSidecarAlerts `field:"optional" json:"alerts" yaml:"alerts"`
	Dashboards *GrafanaSidecarDashboards `field:"optional" json:"dashboards" yaml:"dashboards"`
	Datasources *GrafanaSidecarDatasources `field:"optional" json:"datasources" yaml:"datasources"`
	EnableUniqueFilenames *bool `field:"optional" json:"enableUniqueFilenames" yaml:"enableUniqueFilenames"`
	Image *GrafanaSidecarImage `field:"optional" json:"image" yaml:"image"`
	ImagePullPolicy *string `field:"optional" json:"imagePullPolicy" yaml:"imagePullPolicy"`
	LivenessProbe interface{} `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	Notifiers *GrafanaSidecarNotifiers `field:"optional" json:"notifiers" yaml:"notifiers"`
	Plugins *GrafanaSidecarPlugins `field:"optional" json:"plugins" yaml:"plugins"`
	ReadinessProbe interface{} `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	Resources interface{} `field:"optional" json:"resources" yaml:"resources"`
	SecurityContext interface{} `field:"optional" json:"securityContext" yaml:"securityContext"`
}

