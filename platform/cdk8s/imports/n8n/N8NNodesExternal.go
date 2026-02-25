package n8n


// External npm packages to install and allow in the Code node.
type N8NNodesExternal struct {
	// Allow all external npm packages.
	AllowAll *bool `field:"required" json:"allowAll" yaml:"allowAll"`
	// List of npm package names and versions (e.g., "package-name@1.0.0").
	Packages *[]*string `field:"required" json:"packages" yaml:"packages"`
	// Whether to reinstall missing packages.
	//
	// For more information, see https://docs.n8n.io/integrations/community-nodes/troubleshooting/#error-missing-packages
	ReinstallMissingPackages *bool `field:"required" json:"reinstallMissingPackages" yaml:"reinstallMissingPackages"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

