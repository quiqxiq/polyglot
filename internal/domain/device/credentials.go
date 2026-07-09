package device

// Credentials holds the decrypted authentication material for a device.
// Stored encrypted at rest (see the credentials table, migration 000001);
// only port.CredentialVault implementations may produce this type, per
// Polyglot-Architecture.md §2 prinsip 4 ("AI tidak pernah menyentuh
// kredensial mentah").
//
// Extra carries vendor-specific secrets that don't fit Username/Password
// (e.g. SNMP community string for zteolt, x-api-key for genieacs). These
// are merged into Target.Extra by Device.ToTarget, so drivers read them
// from the same map as non-sensitive params — the driver does not know
// (and should not care) which fields came from the encrypted blob.
type Credentials struct {
	Username string
	Password string
	Extra    map[string]string
}
