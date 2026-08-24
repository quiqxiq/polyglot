// Package importer menyediakan parsing & penulisan data pelanggan dalam
// format CSV/XLSX (kompatibel ekspor Mikhmon / Data_Pelanggan_*.xlsx)
// serta upsert ke DB dan sinkronisasi live-router.
package importer

import (
	"strconv"
	"strings"
)

// Row adalah satu baris pelanggan pada file impor/ekspor.
type Row struct {
	CustomerCode string
	Name         string
	Phone        string
	Email        string
	Address      string
	Latitude     *float64
	Longitude    *float64
	ServiceType  string // PPPOE | HOTSPOT (default PPPOE)
	DeviceName   string // nama router/server tujuan (mis. "JAYA ABADI")
	Username     string
	Password     string
	PlanName     string
	Price        float64
	RateLimit    string // "5M/5M" — opsional
	Status       string // ACTIVE | ISOLATED | SUSPENDED | TERMINATED
	LocalAddress string
	RemoteAddr   string
	ParentQueue  string

	RowNumber int // untuk pesan error ramah
}

// Header alias yang diterima per kolom kanonik.
var headerAliases = map[string][]string{
	"customer_code": {"id_pelanggan", "customer_code", "kode", "kode_pelanggan"},
	"name":          {"nama", "name", "pelanggan"},
	"phone":         {"nomor_telepon", "phone", "no_hp", "telepon", "wa"},
	"email":         {"email"},
	"address":       {"alamat", "address"},
	"latitude":      {"latitude", "lat"},
	"longitude":     {"longitude", "lng", "lon", "long"},
	"service_type":  {"tipe", "service_type", "jenis"},
	"device_name":   {"server", "device_name", "router"},
	"username":      {"username", "user"},
	"password":      {"password", "pass"},
	"plan_name":     {"paket", "plan_name", "profile", "profil"},
	"price":         {"harga", "price", "biaya"},
	"rate_limit":    {"rate_limit", "ratelimit"},
	"status":        {"status"},
	"local_address": {"local_address", "local"},
	"remote_addr":   {"remote_address", "remote_addr", "remote"},
	"parent_queue":  {"parent_queue", "queue"},
}

// mapHeaders memetakan indeks kolom → field kanonik dari baris header.
func mapHeaders(cells []string) map[int]string {
	out := make(map[int]string, len(cells))
	for i, raw := range cells {
		key := strings.ToLower(strings.TrimSpace(raw))
		key = strings.ReplaceAll(key, " ", "_")
		for canonical, aliases := range headerAliases {
			for _, a := range aliases {
				if key == a {
					out[i] = canonical
					break
				}
			}
		}
	}
	return out
}

// buildRow mengubah sel satu baris menjadi Row berdasarkan pemetaan header.
func buildRow(m map[int]string, cells []string, rowNo int) Row {
	var r Row
	r.RowNumber = rowNo
	get := func(field string) string {
		for i, f := range m {
			if f == field && i < len(cells) {
				return strings.TrimSpace(cells[i])
			}
		}
		return ""
	}
	r.CustomerCode = get("customer_code")
	r.Name = get("name")
	r.Phone = get("phone")
	r.Email = get("email")
	r.Address = get("address")
	r.ServiceType = strings.ToUpper(get("service_type"))
	if r.ServiceType == "" {
		r.ServiceType = "PPPOE"
	}
	r.DeviceName = get("device_name")
	r.Username = get("username")
	r.Password = get("password")
	r.PlanName = get("plan_name")
	r.RateLimit = get("rate_limit")
	r.Status = strings.ToUpper(get("status"))
	if r.Status == "" {
		r.Status = "ACTIVE"
	}
	r.LocalAddress = get("local_address")
	r.RemoteAddr = get("remote_addr")
	r.ParentQueue = get("parent_queue")

	if v := get("latitude"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			r.Latitude = &f
		}
	}
	if v := get("longitude"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			r.Longitude = &f
		}
	}
	if v := get("price"); v != "" {
		clean := strings.NewReplacer("Rp", "", ".", "", ",", ".", " ", "").Replace(v)
		if f, err := strconv.ParseFloat(strings.TrimSuffix(clean, "."), 64); err == nil {
			r.Price = f
		}
	}
	return r
}

// ValidateRows mengembalikan daftar kesalahan per baris (tanpa henti di awal).
func ValidateRows(rows []Row) []error {
	var errs []error
	for _, r := range rows {
		line := fmtLine(r.RowNumber)
		if strings.TrimSpace(r.Name) == "" {
			errs = append(errs, errf("%s nama wajib diisi", line))
		}
		if strings.TrimSpace(r.Phone) == "" {
			errs = append(errs, errf("%s nomor telepon wajib diisi", line))
		}
		if strings.TrimSpace(r.Address) == "" {
			errs = append(errs, errf("%s alamat wajib diisi", line))
		}
		if strings.TrimSpace(r.Username) != "" && strings.TrimSpace(r.PlanName) == "" {
			errs = append(errs, errf("%s paket wajib bila username diisi", line))
		}
		switch r.Status {
		case "", "ACTIVE", "PENDING", "ISOLATED", "SUSPENDED", "TERMINATED":
			// kosong → default ACTIVE saat impor
		default:
			errs = append(errs, errf("%s status tidak dikenal: %q", line, r.Status))
		}
	}
	return errs
}
