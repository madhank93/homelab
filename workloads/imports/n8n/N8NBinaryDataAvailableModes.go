package n8n


type N8NBinaryDataAvailableModes string

const (
	// filesystem.
	N8NBinaryDataAvailableModes_FILESYSTEM N8NBinaryDataAvailableModes = "FILESYSTEM"
	// s3.
	N8NBinaryDataAvailableModes_S3 N8NBinaryDataAvailableModes = "S3"
)

