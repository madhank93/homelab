package headlamp


// Protocol (TCP/UDP/SCTP).
//
// Defaults to TCP.
// Default: TCP.
//
type HeadlampServiceExtraServicePortsProtocol string

const (
	// TCP.
	HeadlampServiceExtraServicePortsProtocol_TCP HeadlampServiceExtraServicePortsProtocol = "TCP"
	// UDP.
	HeadlampServiceExtraServicePortsProtocol_UDP HeadlampServiceExtraServicePortsProtocol = "UDP"
	// SCTP.
	HeadlampServiceExtraServicePortsProtocol_SCTP HeadlampServiceExtraServicePortsProtocol = "SCTP"
)

