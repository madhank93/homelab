package grafana


type GrafanaSidecarPlugins struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Env interface{} `field:"optional" json:"env" yaml:"env"`
	ExtraMounts *[]interface{} `field:"optional" json:"extraMounts" yaml:"extraMounts"`
	InitPlugins *bool `field:"optional" json:"initPlugins" yaml:"initPlugins"`
	Label *string `field:"optional" json:"label" yaml:"label"`
	LabelValue *string `field:"optional" json:"labelValue" yaml:"labelValue"`
	ReloadUrl *string `field:"optional" json:"reloadUrl" yaml:"reloadUrl"`
	Resource *string `field:"optional" json:"resource" yaml:"resource"`
	ResourceName *string `field:"optional" json:"resourceName" yaml:"resourceName"`
	Script *string `field:"optional" json:"script" yaml:"script"`
	SearchNamespace interface{} `field:"optional" json:"searchNamespace" yaml:"searchNamespace"`
	SizeLimit interface{} `field:"optional" json:"sizeLimit" yaml:"sizeLimit"`
	SkipReload *bool `field:"optional" json:"skipReload" yaml:"skipReload"`
	WatchMethod *string `field:"optional" json:"watchMethod" yaml:"watchMethod"`
}

