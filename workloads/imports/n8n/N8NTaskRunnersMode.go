package n8n


// Use `internal` to use internal task runner, or use `external` to have external sidecar task runner.
type N8NTaskRunnersMode string

const (
	// internal.
	N8NTaskRunnersMode_INTERNAL N8NTaskRunnersMode = "INTERNAL"
	// external.
	N8NTaskRunnersMode_EXTERNAL N8NTaskRunnersMode = "EXTERNAL"
)

