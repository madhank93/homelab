package n8n


// Type of database to use.
//
// Valid values 'sqlite' | 'postgresdb'.
type N8NDbType string

const (
	// sqlite.
	N8NDbType_SQLITE N8NDbType = "SQLITE"
	// postgresdb.
	N8NDbType_POSTGRESDB N8NDbType = "POSTGRESDB"
)

