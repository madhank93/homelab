package n8n


type N8NLogScopes string

const (
	// concurrency.
	N8NLogScopes_CONCURRENCY N8NLogScopes = "CONCURRENCY"
	// external-secrets.
	N8NLogScopes_EXTERNAL_HYPHEN_SECRETS N8NLogScopes = "EXTERNAL_HYPHEN_SECRETS"
	// license.
	N8NLogScopes_LICENSE N8NLogScopes = "LICENSE"
	// multi-main-setup.
	N8NLogScopes_MULTI_HYPHEN_MAIN_HYPHEN_SETUP N8NLogScopes = "MULTI_HYPHEN_MAIN_HYPHEN_SETUP"
	// pubsub.
	N8NLogScopes_PUBSUB N8NLogScopes = "PUBSUB"
	// redis.
	N8NLogScopes_REDIS N8NLogScopes = "REDIS"
	// scaling.
	N8NLogScopes_SCALING N8NLogScopes = "SCALING"
	// waiting-executions.
	N8NLogScopes_WAITING_HYPHEN_EXECUTIONS N8NLogScopes = "WAITING_HYPHEN_EXECUTIONS"
)

