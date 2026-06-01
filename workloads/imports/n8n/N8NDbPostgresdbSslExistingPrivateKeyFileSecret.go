package n8n


// The PostgreSQL existing SSL private key file secret.
type N8NDbPostgresdbSslExistingPrivateKeyFileSecret struct {
	// The key of the SSL private key in the existing secret.
	Key *string `field:"required" json:"key" yaml:"key"`
	// The name of the existing secret.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

