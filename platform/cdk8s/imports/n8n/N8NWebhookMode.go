package n8n


// Use `regular` to use main node as webhook node, or use `queue` to have webhook nodes.
type N8NWebhookMode string

const (
	// regular.
	N8NWebhookMode_REGULAR N8NWebhookMode = "REGULAR"
	// queue.
	N8NWebhookMode_QUEUE N8NWebhookMode = "QUEUE"
)

