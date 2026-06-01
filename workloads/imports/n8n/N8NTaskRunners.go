package n8n


type N8NTaskRunners struct {
	Broker *N8NTaskRunnersBroker `field:"required" json:"broker" yaml:"broker"`
	// The maximum concurrency for the task.
	MaxConcurrency *float64 `field:"required" json:"maxConcurrency" yaml:"maxConcurrency"`
	// Use `internal` to use internal task runner, or use `external` to have external sidecar task runner.
	Mode N8NTaskRunnersMode `field:"required" json:"mode" yaml:"mode"`
	// The heartbeat interval for the task in seconds.
	TaskHeartbeatInterval *float64 `field:"required" json:"taskHeartbeatInterval" yaml:"taskHeartbeatInterval"`
	// The timeout for the task in seconds.
	TaskTimeout *float64 `field:"required" json:"taskTimeout" yaml:"taskTimeout"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	External *N8NTaskRunnersExternal `field:"optional" json:"external" yaml:"external"`
}

