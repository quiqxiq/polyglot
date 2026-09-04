package ppp

// ActiveSession represents one row returned by /ppp/active/print.
// This is a read-only monitoring resource — the only mutation is
// removing (kicking/disconnecting) an active session.
type ActiveSession struct {
	RosID     string
	Name      string
	Service   string
	CallerID  string
	Address   string
	Encoding  string
	SessionID string
	Radius    bool
	Profile   string
	Uptime    string
}

// PPPoESecret represents one row returned by /ppp/secret/print. Only fields
// that are stable across RouterOS versions and genuinely useful to the
// application are exposed.
type PPPoESecret struct {
	RosID         string // RouterOS internal .id, needed for set/remove
	Name          string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
	LastLoggedOut string
	CallerID      string
}

// Profile represents one row returned by /ppp/profile/print.
type Profile struct {
	RosID          string
	Name           string
	RateLimit      string
	LocalAddress   string
	RemoteAddress  string
	DNSServer      string
	ParentQueue    string
	AddressList    string
	Comment        string
	SharedUsers    string
	OnlyOne        string
	UseMPLS        string
	UseCompression string
	UseEncryption  string
	ChangeTCPMSS   string
	BridgeLearning string
	OnUp           string
	OnDown         string
}

// FilterInactiveSecrets returns secrets that do not appear in the active sessions list.
func FilterInactiveSecrets(secrets []PPPoESecret, active []ActiveSession) []PPPoESecret {
	activeMap := make(map[string]bool, len(active))
	for _, s := range active {
		activeMap[s.Name] = true
	}
	inactive := make([]PPPoESecret, 0)
	for _, s := range secrets {
		if !activeMap[s.Name] {
			inactive = append(inactive, s)
		}
	}
	return inactive
}

// EnrichActiveSessionsWithProfiles enriches PPP active sessions with profile from secrets.
func EnrichActiveSessionsWithProfiles(active []ActiveSession, secrets []PPPoESecret) []ActiveSession {
	secretMap := make(map[string]string, len(secrets))
	for _, s := range secrets {
		if s.Profile != "" {
			secretMap[s.Name] = s.Profile
		}
	}
	for i := range active {
		if active[i].Profile == "" {
			if prof, ok := secretMap[active[i].Name]; ok {
				active[i].Profile = prof
			}
		}
	}
	return active
}
