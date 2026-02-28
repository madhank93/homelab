package n8n


type N8NDbSqlite struct {
	// SQLite database file name.
	Database *string `field:"required" json:"database" yaml:"database"`
	// SQLite database pool size.
	//
	// Set to `0` to disable pooling.
	PoolSize *float64 `field:"required" json:"poolSize" yaml:"poolSize"`
	// Runs VACUUM operation on startup to rebuild the database.
	//
	// Reduces file size and optimizes indexes. This is a long running blocking operation and increases start-up time.
	Vacuum *bool `field:"required" json:"vacuum" yaml:"vacuum"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

