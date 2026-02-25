package n8n


type N8NDbPostgresdb struct {
	// Postgres connection timeout (ms).
	ConnectionTimeout *float64 `field:"required" json:"connectionTimeout" yaml:"connectionTimeout"`
	// Amount of time before an idle connection is eligible for eviction for being idle.
	IdleConnectionTimeout *float64 `field:"required" json:"idleConnectionTimeout" yaml:"idleConnectionTimeout"`
	// Control how many parallel open Postgres connections n8n should have.
	//
	// Increasing it may help with resource utilization, but too many connections may degrade performance.
	PoolSize *float64 `field:"required" json:"poolSize" yaml:"poolSize"`
	// The PostgreSQL schema.
	Schema *string `field:"required" json:"schema" yaml:"schema"`
	// The PostgreSQL connection SSL settings.
	//
	// Find more information from here: https://docs.n8n.io/hosting/configuration/supported-databases-settings/#postgresdb
	Ssl *N8NDbPostgresdbSsl `field:"required" json:"ssl" yaml:"ssl"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

