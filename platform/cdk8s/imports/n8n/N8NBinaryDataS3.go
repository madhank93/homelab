package n8n


// S3-compatible external storage configurations.
//
// For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/external-data-storage/
type N8NBinaryDataS3 struct {
	// Access key in S3-compatible external storage.
	AccessKey *string `field:"required" json:"accessKey" yaml:"accessKey"`
	// Access secret in S3-compatible external storage.
	AccessSecret *string `field:"required" json:"accessSecret" yaml:"accessSecret"`
	// Name of the n8n bucket in S3-compatible external storage.
	BucketName *string `field:"required" json:"bucketName" yaml:"bucketName"`
	// Region of the S3-compatible external storage bucket.
	//
	// For example, us-east-1.
	BucketRegion *string `field:"required" json:"bucketRegion" yaml:"bucketRegion"`
	// This is for setting up the s3 file storage existing secret.
	//
	// Must contain access-key-id and secret-access-key keys.
	ExistingSecret *string `field:"required" json:"existingSecret" yaml:"existingSecret"`
	// Host of the n8n bucket in S3-compatible external storage.
	//
	// For example, s3.us-east-1.amazonaws.com
	Host *string `field:"required" json:"host" yaml:"host"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

