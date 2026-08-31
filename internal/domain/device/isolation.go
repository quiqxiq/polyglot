package device

// IsolationConfig defines the configuration for router isolation provisioning.
type IsolationConfig struct {
	PPPoEProfileName    string   `json:"pppoe_profile_name"`
	HotspotProfileName  string   `json:"hotspot_profile_name"`
	AddressListName     string   `json:"address_list_name"`
	RateLimit           string   `json:"rate_limit"`
	LocalAddress        string   `json:"local_address"`
	RemoteAddressPool   string   `json:"remote_address_pool"`
	RedirectIP          string   `json:"redirect_ip"`
	RedirectPort        int      `json:"redirect_port"`
	NATRedirectEnabled  bool     `json:"nat_redirect_enabled"`
	PPPoERedirectURL    string   `json:"pppoe_redirect_url"`
	HotspotRedirectURL  string   `json:"hotspot_redirect_url"`
	WalledGardenDomains []string `json:"walled_garden_domains"`
}

// DefaultIsolationConfig returns production default isolation settings.
func DefaultIsolationConfig() IsolationConfig {
	return IsolationConfig{
		PPPoEProfileName:   "ISOLIR",
		HotspotProfileName: "ISOLIR",
		AddressListName:    "ISOLIR_USERS",
		RateLimit:          "10M/10M",
		LocalAddress:       "10.100.0.1",
		RemoteAddressPool:  "pool-isolir",
		RedirectIP:         "192.168.233.195",
		RedirectPort:       5176,
		NATRedirectEnabled: true,
		PPPoERedirectURL:   "http://isolir.isp.net:5176/portal/isolate/pppoe",
		HotspotRedirectURL: "http://isolir.isp.net:5176/portal/isolate/hotspot",
		WalledGardenDomains: []string{
			"tripay.co.id",
			"api.tripay.co.id",
			"midtrans.com",
			"app.midtrans.com",
			"whatsapp.com",
			"isolir.isp.net",
			"bayar.isp.net",
			"portal.polyglot.test",
		},
	}
}

// IsolationStatus represents the live isolation status on a router.
type IsolationStatus struct {
	PPPoEProfileExists   bool            `json:"pppoe_profile_exists"`
	HotspotProfileExists bool            `json:"hotspot_profile_exists"`
	AddressListExists    bool            `json:"address_list_exists"`
	NATRedirectExists    bool            `json:"nat_redirect_exists"`
	IsolatedUsersCount   int             `json:"isolated_users_count"`
	Config               IsolationConfig `json:"config"`
}

// RouterIntegrationScripts holds generated RouterOS scripts for webhook integration.
type RouterIntegrationScripts struct {
	PPPOnUpScript         string `json:"ppp_on_up_script"`
	PPPOnDownScript       string `json:"ppp_on_down_script"`
	HotspotOnLoginScript  string `json:"hotspot_on_login_script"`
	HotspotOnLogoutScript string `json:"hotspot_on_logout_script"`
	WebhookToken          string `json:"webhook_token"`
}
