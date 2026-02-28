package n8n


type N8NLogFile struct {
	// Location of the log files inside `~/.n8n`. Only for `file` log output.
	Location *string `field:"required" json:"location" yaml:"location"`
	// Max number of log files to keep, or max number of days to keep logs for.
	//
	// Once the limit is reached, the oldest log files will be rotated out. If using days, append a `d` suffix. Only for `file` log output.
	Maxcount *string `field:"required" json:"maxcount" yaml:"maxcount"`
	// The maximum size (in MB) for each log file.
	//
	// By default, n8n uses 16 MB.
	Maxsize *float64 `field:"required" json:"maxsize" yaml:"maxsize"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

