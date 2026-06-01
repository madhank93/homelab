package n8n


// The PostgreSQL connection SSL settings.
//
// Find more information from here: https://docs.n8n.io/hosting/configuration/supported-databases-settings/#postgresdb
type N8NDbPostgresdbSsl struct {
	// The PostgreSQL base64 encoded version of SSL certificate file content.
	Base64EncodedCertFile *string `field:"required" json:"base64EncodedCertFile" yaml:"base64EncodedCertFile"`
	// The PostgreSQL base64 encoded version of SSL certificate authority file content.
	Base64EncodedCertificateAuthorityFile *string `field:"required" json:"base64EncodedCertificateAuthorityFile" yaml:"base64EncodedCertificateAuthorityFile"`
	// The PostgreSQL base64 encoded version of SSL private key file content.
	Base64EncodedPrivateKeyFile *string `field:"required" json:"base64EncodedPrivateKeyFile" yaml:"base64EncodedPrivateKeyFile"`
	// Whether to enable SSL.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// The PostgreSQL existing certificate file secret.
	ExistingCertFileSecret *N8NDbPostgresdbSslExistingCertFileSecret `field:"required" json:"existingCertFileSecret" yaml:"existingCertFileSecret"`
	// The PostgreSQL existing certificate authority file secret.
	ExistingCertificateAuthorityFileSecret *N8NDbPostgresdbSslExistingCertificateAuthorityFileSecret `field:"required" json:"existingCertificateAuthorityFileSecret" yaml:"existingCertificateAuthorityFileSecret"`
	// The PostgreSQL existing SSL private key file secret.
	ExistingPrivateKeyFileSecret *N8NDbPostgresdbSslExistingPrivateKeyFileSecret `field:"required" json:"existingPrivateKeyFileSecret" yaml:"existingPrivateKeyFileSecret"`
	// If n8n should reject unauthorized SSL connections (true) or not (false).
	RejectUnauthorized *bool `field:"required" json:"rejectUnauthorized" yaml:"rejectUnauthorized"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

