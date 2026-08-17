package headlamp


// Scheme for probes (HTTP or HTTPS).
//
// Set to HTTPS when TLS is enabled at the backend server.
type HeadlampProbesScheme string

const (
	// HTTP.
	HeadlampProbesScheme_HTTP HeadlampProbesScheme = "HTTP"
	// HTTPS.
	HeadlampProbesScheme_HTTPS HeadlampProbesScheme = "HTTPS"
)

