package zteolt

// telnetClient is the raw Telnet connection used for ZTE OLT provisioning
// commands (there is no scrapligo platform definition for ZTE, see
// TECH-STACK-DAN-PERSIAPAN.md §8 — hence a raw client here, not scrapligo).
// TODO: replace the empty struct with an actual net.Conn-based Telnet
// session wrapper (dial, prompt handling, read/write).
type telnetClient struct{}

// close releases the underlying Telnet connection.
func (t *telnetClient) close() error {
	return nil
}
