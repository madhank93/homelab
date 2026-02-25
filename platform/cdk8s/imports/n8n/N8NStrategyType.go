package n8n


type N8NStrategyType string

const (
	// RollingUpdate.
	N8NStrategyType_ROLLING_UPDATE N8NStrategyType = "ROLLING_UPDATE"
	// Recreate.
	N8NStrategyType_RECREATE N8NStrategyType = "RECREATE"
)

