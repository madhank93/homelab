package n8n


// The log output level.
//
// The available options are (from lowest to highest level) are error, warn, info, and debug. The default value is info. You can learn more about these options [here](https://docs.n8n.io/hosting/logging-monitoring/logging/#log-levels).
type N8NLogLevel string

const (
	// error.
	N8NLogLevel_ERROR N8NLogLevel = "ERROR"
	// warn.
	N8NLogLevel_WARN N8NLogLevel = "WARN"
	// info.
	N8NLogLevel_INFO N8NLogLevel = "INFO"
	// debug.
	N8NLogLevel_DEBUG N8NLogLevel = "DEBUG"
)

