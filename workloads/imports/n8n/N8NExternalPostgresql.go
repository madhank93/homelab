package n8n


// External PostgreSQL parameters.
type N8NExternalPostgresql struct {
	// The name of the external PostgreSQL database.
	//
	// For more information: https://docs.n8n.io/hosting/configuration/supported-databases-settings/#required-permissions
	Database *string `field:"required" json:"database" yaml:"database"`
	// The name of an existing secret with PostgreSQL (must contain key `postgres-password`) and credentials.
	//
	// When it's set, the `externalPostgresql.password` parameter is ignored
	ExistingSecret *string `field:"required" json:"existingSecret" yaml:"existingSecret"`
	// External PostgreSQL server host.
	Host *string `field:"required" json:"host" yaml:"host"`
	// External PostgreSQL password.
	Password *string `field:"required" json:"password" yaml:"password"`
	// External PostgreSQL server port.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// External PostgreSQL username.
	Username *string `field:"required" json:"username" yaml:"username"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

