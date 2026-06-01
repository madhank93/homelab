package n8n


// Access mode for persistence.
type N8NMainPersistenceAccessMode string

const (
	// ReadWriteOnce.
	N8NMainPersistenceAccessMode_READ_WRITE_ONCE N8NMainPersistenceAccessMode = "READ_WRITE_ONCE"
	// ReadWriteMany.
	N8NMainPersistenceAccessMode_READ_WRITE_MANY N8NMainPersistenceAccessMode = "READ_WRITE_MANY"
)

