package headlamp


// Experimental/alpha Cluster Inventory configuration.
type HeadlampConfigClusterInventory struct {
	// Experimental/alpha Cluster Inventory access providers config.
	AccessProvidersConfig interface{} `field:"optional" json:"accessProvidersConfig" yaml:"accessProvidersConfig"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Enable experimental/alpha Cluster Inventory discovery.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Kubernetes label selector used to filter experimental/alpha ClusterProfile resources.
	LabelSelector *string `field:"optional" json:"labelSelector" yaml:"labelSelector"`
	// Override the experimental/alpha Cluster Inventory no-CRD cache TTL.
	//
	// Empty uses the Headlamp default.
	NoCrdCacheTtl *string `field:"optional" json:"noCrdCacheTtl" yaml:"noCrdCacheTtl"`
	// Kubernetes image volumes that provide experimental/alpha Cluster Inventory access provider binaries.
	Plugins *[]*HeadlampConfigClusterInventoryPlugins `field:"optional" json:"plugins" yaml:"plugins"`
	// Override the experimental/alpha Cluster Inventory root reconcile interval.
	//
	// Empty uses the Headlamp default.
	RootReconcileInterval *string `field:"optional" json:"rootReconcileInterval" yaml:"rootReconcileInterval"`
}

