package n8n


// Database logging level.
//
// Requires `maxQueryExecutionTime` to be higher than `0`. Valid values 'query' | 'error' | 'schema' | 'warn' | 'info' | 'log' | 'all'
type N8NDbLoggingOptions string

const (
	// query.
	N8NDbLoggingOptions_QUERY N8NDbLoggingOptions = "QUERY"
	// error.
	N8NDbLoggingOptions_ERROR N8NDbLoggingOptions = "ERROR"
	// schema.
	N8NDbLoggingOptions_SCHEMA N8NDbLoggingOptions = "SCHEMA"
	// warn.
	N8NDbLoggingOptions_WARN N8NDbLoggingOptions = "WARN"
	// info.
	N8NDbLoggingOptions_INFO N8NDbLoggingOptions = "INFO"
	// log.
	N8NDbLoggingOptions_LOG N8NDbLoggingOptions = "LOG"
	// all.
	N8NDbLoggingOptions_ALL N8NDbLoggingOptions = "ALL"
)

