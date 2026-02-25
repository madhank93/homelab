package n8n


// Access mode for persistence.
type N8NWorkerPersistenceAccessMode string

const (
	// ReadWriteOnce.
	N8NWorkerPersistenceAccessMode_READ_WRITE_ONCE N8NWorkerPersistenceAccessMode = "READ_WRITE_ONCE"
	// ReadWriteMany.
	N8NWorkerPersistenceAccessMode_READ_WRITE_MANY N8NWorkerPersistenceAccessMode = "READ_WRITE_MANY"
)

