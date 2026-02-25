package n8n


// The default binary data mode.
//
// default keeps binary data in memory. Set to filesystem to use the filesystem, or s3 to AWS S3. Note that binary data pruning operates on the active binary data mode. For example, if your instance stored data in S3, and you later switched to filesystem mode, n8n only prunes binary data in the filesystem. This may change in future. Valid values are 'default' | 'filesystem' | 's3'. For more information, see https://docs.n8n.io/hosting/configuration/environment-variables/binary-data/
type N8NBinaryDataMode string

const (
	// default.
	N8NBinaryDataMode_DEFAULT N8NBinaryDataMode = "DEFAULT"
	// filesystem.
	N8NBinaryDataMode_FILESYSTEM N8NBinaryDataMode = "FILESYSTEM"
	// s3.
	N8NBinaryDataMode_S3 N8NBinaryDataMode = "S3"
)

