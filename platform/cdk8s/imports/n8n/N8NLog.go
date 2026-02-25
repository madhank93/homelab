package n8n


// n8n log configurations.
type N8NLog struct {
	File *N8NLogFile `field:"required" json:"file" yaml:"file"`
	// The log output level.
	//
	// The available options are (from lowest to highest level) are error, warn, info, and debug. The default value is info. You can learn more about these options [here](https://docs.n8n.io/hosting/logging-monitoring/logging/#log-levels).
	Level N8NLogLevel `field:"required" json:"level" yaml:"level"`
	// Where to output logs to.
	//
	// Options are: `console` or `file` or both.
	Output *[]N8NLogOutput `field:"required" json:"output" yaml:"output"`
	// Scopes to filter logs by.
	//
	// Nothing is filtered by default. Supported log scopes: concurrency, external-secrets, license, multi-main-setup, pubsub, redis, scaling, waiting-executions
	Scopes *[]N8NLogScopes `field:"required" json:"scopes" yaml:"scopes"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

