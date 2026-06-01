package n8n


// Use `regular` to use main node as executer, or use `queue` to have worker nodes.
type N8NWorkerMode string

const (
	// regular.
	N8NWorkerMode_REGULAR N8NWorkerMode = "REGULAR"
	// queue.
	N8NWorkerMode_QUEUE N8NWorkerMode = "QUEUE"
)

