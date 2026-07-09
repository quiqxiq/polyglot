package device

import (
	"time"
)

// Target describes how to reach and authenticate to a device — the minimal
// connection parameters a DeviceDriver needs. Distinct from Device below:
// Device is the stored inventory record (name, vendor, tags, ...); Target
// is just what NewDriver(ctx, target) needs to connect.
// Extra holds vendor-specific parameters that don't fit the common fields
// (e.g. the SNMP community string for zteolt, or use_tls for genieacs).
type Target struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
	Extra    map[string]string
}

// Device is the stored inventory record for a network device, mapped to the
// `devices` table per Polyglot-Architecture.md §7.2. It holds non-sensitive
// connection parameters and vendor metadata; credentials live separately in
// the encrypted credentials table (see Credentials + port.CredentialVault).
//
// Named Device (the type), not NewDevice — the constructor pattern for this
// entity is still TODO (validation, defaults); callers currently construct
// the struct literal directly, which is fine for the fields it has.
type Device struct {
	ID             string
	Name           string
	Vendor         string
	DriverType     string
	Host           string
	Port           int
	TimeoutMS      int
	PollIntervalMS int
	Extra          map[string]string
	Tags           []string
	Enabled        bool
}

// ToTarget merges a Device inventory record with decrypted Credentials into
// a Target suitable for passing to a vendor's NewDriver(ctx, target).
// Non-sensitive vendor params from Device.Extra (e.g. use_tls, device_id for
// genieacs) and sensitive ones from Credentials.Extra (e.g. api_key,
// community string) are merged into a single Extra map — the driver sees one
// uniform map regardless of where each field physically lives. Credential
// fields win on key conflict, since they are the authoritative source for
// secrets.
func (d Device) ToTarget(c Credentials) Target {
	extra := make(map[string]string, len(d.Extra)+len(c.Extra))
	for k, v := range d.Extra {
		extra[k] = v
	}
	for k, v := range c.Extra {
		extra[k] = v
	}
	return Target{
		Host:     d.Host,
		Port:     d.Port,
		Username: c.Username,
		Password: c.Password,
		Timeout:  time.Duration(d.TimeoutMS) * time.Millisecond,
		Extra:    extra,
	}
}
