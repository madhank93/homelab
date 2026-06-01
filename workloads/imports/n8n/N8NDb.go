package n8n


// n8n database configurations.
type N8NDb struct {
	Logging *N8NDbLogging `field:"required" json:"logging" yaml:"logging"`
	Postgresdb *N8NDbPostgresdb `field:"required" json:"postgresdb" yaml:"postgresdb"`
	Sqlite *N8NDbSqlite `field:"required" json:"sqlite" yaml:"sqlite"`
	// Prefix to use for table names.
	TablePrefix *string `field:"required" json:"tablePrefix" yaml:"tablePrefix"`
	// Type of database to use.
	//
	// Valid values 'sqlite' | 'postgresdb'.
	Type N8NDbType `field:"required" json:"type" yaml:"type"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

