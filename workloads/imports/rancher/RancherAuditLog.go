package rancher


type RancherAuditLog struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// auditLog.destination must be either 'sidecar' or 'hostPath'.
	Destination RancherAuditLogDestination `field:"optional" json:"destination" yaml:"destination"`
	// auditLog.enabled must be a boolean.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// auditLog.level must be a number 0-3; 0 to disable, 3 for most verbose.
	Level RancherAuditLogLevel `field:"optional" json:"level" yaml:"level"`
}

