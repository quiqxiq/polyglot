package ont

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// ACSDevice adalah cache satu device GenieACS (CPE/ONT) untuk pencocokan.
type ACSDevice struct {
	ID              string          `json:"id"`
	TenantID        string          `json:"tenant_id"`
	ACSDeviceID     string          `json:"acs_device_id"`
	OUI             string          `json:"oui,omitempty"`
	ProductClass    string          `json:"product_class,omitempty"`
	SerialNumber    string          `json:"serial_number,omitempty"`
	SoftwareVersion string          `json:"software_version,omitempty"`
	LastInformAt    *time.Time      `json:"last_inform_at,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	MACCandidates   []string        `json:"mac_candidates,omitempty"`
	Raw             json.RawMessage `json:"raw,omitempty"`
	MatchedONTID    *string         `json:"matched_ont_id,omitempty"`
	MatchSource     string          `json:"match_source,omitempty"`
	FirstSeenAt     time.Time       `json:"first_seen_at"`
	LastSeenAt      time.Time       `json:"last_seen_at"`
}

// ACSPreset adalah preset GenieACS yang dikelola Polyglot (bootstrap ACS).
type ACSPreset struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	Name           string          `json:"name"`
	Precondition   json.RawMessage `json:"precondition,omitempty"`
	Configurations json.RawMessage `json:"configurations"`
	Enabled        bool            `json:"enabled"`
	LastSyncedAt   *time.Time      `json:"last_synced_at,omitempty"`
	LastError      string          `json:"last_error,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// TelemetryCatalog memetakan field telemetri ke daftar path parameter TR-069
// (kandidat pertama yang tersedia dipakai; mendukung TR-098, TR-181, dan
// VirtualParameters).
type TelemetryCatalog struct {
	SerialNumber    []string `json:"serial_number"`
	ProductClass    []string `json:"product_class"`
	SoftwareVersion []string `json:"software_version"`
	HardwareVersion []string `json:"hardware_version"`
	Uptime          []string `json:"uptime"`
	WANIP           []string `json:"wan_ip"`
	WANUsername     []string `json:"wan_username"`
	WANStatus       []string `json:"wan_status"`
	SSID24          []string `json:"ssid_24"`
	SSID5           []string `json:"ssid_5"`
	WiFiClients     []string `json:"wifi_clients"`
	RxPowerDBM      []string `json:"rx_power_dbm"`
	TemperatureC    []string `json:"temperature_c"`
}

// DefaultTelemetryCatalog berisi kandidat path dari implementasi lapangan
// (TR-098 dominan untuk ONT EPON; TR-181 & VirtualParameters sebagai fallback).
func DefaultTelemetryCatalog() TelemetryCatalog {
	return TelemetryCatalog{
		SerialNumber: []string{
			"InternetGatewayDevice.DeviceInfo.SerialNumber",
			"Device.DeviceInfo.SerialNumber",
		},
		ProductClass: []string{
			"InternetGatewayDevice.DeviceInfo.ProductClass",
			"Device.DeviceInfo.ProductClass",
		},
		SoftwareVersion: []string{
			"InternetGatewayDevice.DeviceInfo.SoftwareVersion",
			"Device.DeviceInfo.SoftwareVersion",
		},
		HardwareVersion: []string{
			"InternetGatewayDevice.DeviceInfo.HardwareVersion",
			"Device.DeviceInfo.HardwareVersion",
		},
		Uptime: []string{
			"InternetGatewayDevice.DeviceInfo.UpTime",
			"Device.DeviceInfo.UpTime",
		},
		WANIP: []string{
			"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.ExternalIPAddress",
			"Device.PPP.Interface.1.IPAddress",
		},
		WANUsername: []string{
			"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
			"Device.PPP.Interface.1.Username",
		},
		WANStatus: []string{
			"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.ConnectionStatus",
			"Device.PPP.Interface.1.ConnectionStatus",
		},
		SSID24: []string{
			"InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID",
			"Device.WiFi.SSID.1.SSID",
		},
		SSID5: []string{
			"InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.SSID",
			"Device.WiFi.SSID.2.SSID",
		},
		WiFiClients: []string{
			"InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.TotalAssociations",
			"Device.WiFi.AccessPoint.1.AssociatedDeviceNumberOfEntries",
		},
		RxPowerDBM: []string{
			"InternetGatewayDevice.WANDevice.1.WANPONInterfaceConfig.RXPower",
			"VirtualParameters.RXPower",
			"VirtualParameters.OpticalPower",
			"Device.Optical.Interface.1.OpticalSignalLevel",
		},
		TemperatureC: []string{
			"InternetGatewayDevice.DeviceInfo.Temperature",
			"VirtualParameters.Temperature",
			"Device.DeviceInfo.Temperature",
		},
	}
}

// CatalogFromJSON menggabungkan default dengan override JSON (field yang
// tidak diisi tetap memakai default).
func CatalogFromJSON(raw string) (TelemetryCatalog, error) {
	cat := DefaultTelemetryCatalog()
	if strings.TrimSpace(raw) == "" {
		return cat, nil
	}
	var override TelemetryCatalog
	if err := json.Unmarshal([]byte(raw), &override); err != nil {
		return cat, err
	}
	mergeList := func(base, over []string) []string {
		if len(over) > 0 {
			return over
		}
		return base
	}
	cat.SerialNumber = mergeList(cat.SerialNumber, override.SerialNumber)
	cat.ProductClass = mergeList(cat.ProductClass, override.ProductClass)
	cat.SoftwareVersion = mergeList(cat.SoftwareVersion, override.SoftwareVersion)
	cat.HardwareVersion = mergeList(cat.HardwareVersion, override.HardwareVersion)
	cat.Uptime = mergeList(cat.Uptime, override.Uptime)
	cat.WANIP = mergeList(cat.WANIP, override.WANIP)
	cat.WANUsername = mergeList(cat.WANUsername, override.WANUsername)
	cat.WANStatus = mergeList(cat.WANStatus, override.WANStatus)
	cat.SSID24 = mergeList(cat.SSID24, override.SSID24)
	cat.SSID5 = mergeList(cat.SSID5, override.SSID5)
	cat.WiFiClients = mergeList(cat.WiFiClients, override.WiFiClients)
	cat.RxPowerDBM = mergeList(cat.RxPowerDBM, override.RxPowerDBM)
	cat.TemperatureC = mergeList(cat.TemperatureC, override.TemperatureC)
	return cat, nil
}

// AllPaths mengembalikan seluruh path unik katalog (untuk satu task
// getParameterValues).
func (c TelemetryCatalog) AllPaths() []string {
	seen := map[string]bool{}
	var out []string
	add := func(list []string) {
		for _, p := range list {
			if p != "" && !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	add(c.SerialNumber)
	add(c.ProductClass)
	add(c.SoftwareVersion)
	add(c.HardwareVersion)
	add(c.Uptime)
	add(c.WANIP)
	add(c.WANUsername)
	add(c.WANStatus)
	add(c.SSID24)
	add(c.SSID5)
	add(c.WiFiClients)
	add(c.RxPowerDBM)
	add(c.TemperatureC)
	return out
}

// MatchBySerial mencocokkan serial ONT (dari label/PSB) dengan serial ACS.
func MatchBySerial(ontSerial, acsSerial string) bool {
	a := NormalizeSerial(ontSerial)
	b := NormalizeSerial(acsSerial)
	return a != "" && b != "" && a == b
}

// MatchExactMAC mencocokkan MAC ONT dengan salah satu kandidat MAC ACS (exact).
func MatchExactMAC(ontMAC string, acsMACs []string) (string, bool) {
	want, ok := NormalizeMAC(ontMAC)
	if !ok {
		return "", false
	}
	for _, raw := range acsMACs {
		got, ok := NormalizeMAC(raw)
		if ok && got == want {
			return got, true
		}
	}
	return "", false
}

// MACNeighbor mencocokkan dua MAC yang berbeda maksimum 1 pada oktet terakhir
// (heuristik lapangan untuk ONT/CPE yang memakai blok MAC bersebelahan).
func MACNeighbor(a, b string) bool {
	na, okA := NormalizeMAC(a)
	nb, okB := NormalizeMAC(b)
	if !okA || !okB {
		return false
	}
	if na == nb {
		return true
	}
	la := strings.Split(na, ":")
	lb := strings.Split(nb, ":")
	if len(la) != 6 || len(lb) != 6 {
		return false
	}
	for i := 0; i < 5; i++ {
		if la[i] != lb[i] {
			return false
		}
	}
	x, errA := strconv.ParseInt(la[5], 16, 64)
	y, errB := strconv.ParseInt(lb[5], 16, 64)
	if errA != nil || errB != nil {
		return false
	}
	diff := x - y
	if diff < 0 {
		diff = -diff
	}
	return diff <= 1
}

// BestMACMatch memilih kandidat MAC ACS terbaik untuk MAC ONT.
// Prioritas: exact, lalu neighbor.
func BestMACMatch(ontMAC string, acsMACs []string) (MatchSource, string, bool) {
	if mac, ok := MatchExactMAC(ontMAC, acsMACs); ok {
		return MatchMAC, mac, true
	}
	for _, raw := range acsMACs {
		if MACNeighbor(ontMAC, raw) {
			mac, _ := NormalizeMAC(raw)
			return MatchMACNeighbor, mac, true
		}
	}
	return "", "", false
}
