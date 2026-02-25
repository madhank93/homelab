package n8n


// Configuration for binary data storage.
type N8NBinaryData struct {
	// Path for binary data storage in 'filesystem' mode.
	//
	// If not set, the default path will be used. For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/binary-data/
	LocalStoragePath *string `field:"required" json:"localStoragePath" yaml:"localStoragePath"`
	// The default binary data mode.
	//
	// default keeps binary data in memory. Set to filesystem to use the filesystem, or s3 to AWS S3. Note that binary data pruning operates on the active binary data mode. For example, if your instance stored data in S3, and you later switched to filesystem mode, n8n only prunes binary data in the filesystem. This may change in future. Valid values are 'default' | 'filesystem' | 's3'. For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/binary-data/
	Mode N8NBinaryDataMode `field:"required" json:"mode" yaml:"mode"`
	// S3-compatible external storage configurations.
	//
	// For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/external-data-storage/
	S3 *N8NBinaryDataS3 `field:"required" json:"s3" yaml:"s3"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Available modes of binary data storage.
	//
	// If not set, the default mode will be used. For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/binary-data/
	AvailableModes *[]N8NBinaryDataAvailableModes `field:"optional" json:"availableModes" yaml:"availableModes"`
}

