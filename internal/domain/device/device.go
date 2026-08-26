package device

import (
	"strconv"
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
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	Name           string            `json:"name"`
	Vendor         string            `json:"vendor"`
	DriverType     string            `json:"driver_type"`
	Host           string            `json:"host"`
	Port           int               `json:"port"`
	SSHPort        int               `json:"ssh_port,omitempty"`
	TimeoutMS      int               `json:"timeout_ms"`
	PollIntervalMS int               `json:"poll_interval_ms,omitempty"`
	Extra          map[string]string `json:"extra,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Enabled        bool              `json:"enabled"`
}

// ToTarget merges a Device inventory record with decrypted Credentials into
// a Target suitable for passing to a vendor's NewDriver(ctx, target).
func (d Device) ToTarget(c Credentials) Target {
	extra := make(map[string]string, len(d.Extra)+len(c.Extra)+1)
	for k, v := range d.Extra {
		extra[k] = v
	}
	for k, v := range c.Extra {
		extra[k] = v
	}
	sshPort := d.SSHPort
	if sshPort <= 0 {
		sshPort = 22
	}
	extra["ssh_port"] = strconv.Itoa(sshPort)

	return Target{
		Host:     d.Host,
		Port:     d.Port,
		Username: c.Username,
		Password: c.Password,
		Timeout:  time.Duration(d.TimeoutMS) * time.Millisecond,
		Extra:    extra,
	}
}

// PingConfig parses the ping metrics settings from Extra metadata.
func (d Device) PingConfig() DevicePingConfig {
	cfg := DefaultPingConfig()
	if d.Extra == nil {
		return cfg
	}
	if v, ok := d.Extra["ping_enabled"]; ok {
		cfg.Enabled = v == "true" || v == "1"
	}
	if v, ok := d.Extra["ping_target"]; ok && v != "" {
		cfg.Target = v
	}
	if v, ok := d.Extra["ping_retention_days"]; ok {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			cfg.RetentionDays = days
		}
	}
	return cfg
}

// SetPingConfig stores the ping metrics settings into Extra metadata.
func (d *Device) SetPingConfig(cfg DevicePingConfig) {
	if d.Extra == nil {
		d.Extra = make(map[string]string)
	}
	if cfg.Enabled {
		d.Extra["ping_enabled"] = "true"
	} else {
		d.Extra["ping_enabled"] = "false"
	}
	if cfg.Target != "" {
		d.Extra["ping_target"] = cfg.Target
	} else {
		d.Extra["ping_target"] = "8.8.8.8"
	}
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = 7
	}
	d.Extra["ping_retention_days"] = strconv.Itoa(cfg.RetentionDays)
}
