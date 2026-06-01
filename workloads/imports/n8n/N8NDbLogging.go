package n8n


type N8NDbLogging struct {
	// Whether database logging is enabled.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Only queries that exceed this time (ms) will be logged.
	//
	// Set `0` to disable.
	MaxQueryExecutionTime *float64 `field:"required" json:"maxQueryExecutionTime" yaml:"maxQueryExecutionTime"`
	// Database logging level.
	//
	// Requires `maxQueryExecutionTime` to be higher than `0`. Valid values 'query' | 'error' | 'schema' | 'warn' | 'info' | 'log' | 'all'
	Options N8NDbLoggingOptions `field:"required" json:"options" yaml:"options"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

