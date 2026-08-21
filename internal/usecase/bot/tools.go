package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/ping"
)

// RequestSkillInput adalah argumen untuk pemanggilan tool request_skill.
type RequestSkillInput struct {
	SkillName    string `json:"skill_name" jsonschema:"description=Nama skill yang ingin diminta (contoh: ghaib-network-cs)"`
	ResourcePath string `json:"resource_path,omitempty" jsonschema:"description=Path opsional berkas referensi, contoh: references/troubleshooting-jaringan.md"`
}

// NewRequestSkillTool membuat tool LLM untuk mengambil instruksi SOP skill secara on-demand.
func NewRequestSkillTool(skillProv SkillProvider) llm.Tool {
	return llm.Tool{
		Name:        "request_skill",
		Description: "Mengambil petunjuk teknis, SOP, panduan langkah-demi-langkah, atau berkas referensi dari skill tertentu saat pelanggan menanyakan topik spesifik.",
		InputSchema: RequestSkillInput{},
		Handler: func(ctx context.Context, argsJSON string) (string, error) {
			if skillProv == nil {
				return "Error: SkillProvider tidak tersedia", nil
			}

			var input RequestSkillInput
			if strings.TrimSpace(argsJSON) != "" {
				_ = json.Unmarshal([]byte(argsJSON), &input)
			}

			skillName := strings.TrimSpace(input.SkillName)
			if skillName == "" {
				return "Error: parameter 'skill_name' wajib diisi (misal: 'ghaib-network-cs')", nil
			}

			resPath := strings.TrimSpace(input.ResourcePath)
			if resPath != "" {
				if getter, ok := skillProv.(interface {
					GetSkillResource(ctx context.Context, skillName, path string) ([]byte, string, error)
				}); ok {
					data, mime, err := getter.GetSkillResource(ctx, skillName, resPath)
					if err != nil {
						logger.WithComponent("BotTool").Warnf("Failed to get skill resource %s/%s: %v", skillName, resPath, err)
						return fmt.Sprintf("Gagal memuat referensi '%s' dari skill '%s': %v", resPath, skillName, err), nil
					}
					if strings.HasPrefix(mime, "text/") || strings.Contains(mime, "markdown") || strings.Contains(mime, "json") {
						return fmt.Sprintf("=== REFERENSI: %s (%s) ===\n%s", resPath, skillName, string(data)), nil
					}
					return fmt.Sprintf("Berkas referensi '%s' termuat (tipe biner: %s, %d bytes)", resPath, mime, len(data)), nil
				}
			}

			content, err := skillProv.GetSkillContent(ctx, skillName)
			if err != nil {
				logger.WithComponent("BotTool").Warnf("Failed to get skill content for %s: %v", skillName, err)
				return fmt.Sprintf("Gagal memuat skill '%s': %v", skillName, err), nil
			}

			return fmt.Sprintf("=== SKILL: %s ===\n%s", skillName, content), nil
		},
	}
}

// PingHostInput adalah argumen untuk pemanggilan tool ping_host.
type PingHostInput struct {
	Host  string `json:"host" jsonschema:"description=Alamat IP atau domain target yang ingin dicek latensi/konektivitasnya (contoh: 8.8.8.8, 1.1.1.1)"`
	Count int    `json:"count,omitempty" jsonschema:"description=Jumlah paket ping yang dikirimkan (default 3, maksimum 5)"`
}

// NewPingHostTool membuat tool LLM untuk melakukan uji ping / diagnostik latensi.
func NewPingHostTool() llm.Tool {
	return llm.Tool{
		Name:        "ping_host",
		Description: "Melakukan diagnostik ping jaringan ke alamat IP atau domain target untuk memeriksa latensi, kestabilan koneksi, dan packet loss.",
		InputSchema: PingHostInput{},
		Handler: func(ctx context.Context, argsJSON string) (string, error) {
			var input PingHostInput
			if strings.TrimSpace(argsJSON) != "" {
				_ = json.Unmarshal([]byte(argsJSON), &input)
			}

			host := strings.TrimSpace(input.Host)
			if host == "" {
				return "Error: parameter 'host' wajib diisi (contoh: '8.8.8.8')", nil
			}

			// Validasi keamanan: cegah shell injection
			if strings.ContainsAny(host, " ;&|`$><\n\r") {
				return "Error: format host tidak valid atau mengandung karakter terlarang", nil
			}

			count := input.Count
			if count <= 0 {
				count = 3
			}
			if count > 5 {
				count = 5
			}

			pingCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()

			cmd := exec.CommandContext(pingCtx, "ping", "-c", fmt.Sprintf("%d", count), "-W", "2", host)
			output, err := cmd.CombinedOutput()
			outStr := string(output)

			if err != nil {
				if pingCtx.Err() == context.DeadlineExceeded {
					return fmt.Sprintf("Ping ke %s timeout (RTO / Request Timed Out). Host tidak merespon.", host), nil
				}
				return fmt.Sprintf("Hasil ping ke %s gagal atau tidak terjangkau:\n%s", host, outStr), nil
			}

			// Parse latency
			row := map[string]string{"time": outStr}
			latency, status := ping.ParsePingLatency(row)
			if latency > 0 {
				return fmt.Sprintf("Ping ke %s berhasil: Status=%s, Avg Latency=%dms\nDetail:\n%s", host, status, latency, outStr), nil
			}

			return fmt.Sprintf("Ping ke %s berhasil:\n%s", host, outStr), nil
		},
	}
}

// GetCurrentTimeInput adalah argumen untuk pemanggilan tool get_current_time.
type GetCurrentTimeInput struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"description=Timezone, contoh: Asia/Jakarta, WIB"`
}

// NewGetCurrentTimeTool membuat tool LLM untuk mendapatkan waktu dan kalender saat ini.
func NewGetCurrentTimeTool() llm.Tool {
	return llm.Tool{
		Name:        "get_current_time",
		Description: "Mendapatkan tanggal, hari, bulan, tahun, dan jam saat ini (WIB) untuk memastikan info batas waktu / jatuh tempo tagihan.",
		InputSchema: GetCurrentTimeInput{},
		Handler: func(ctx context.Context, argsJSON string) (string, error) {
			loc, err := time.LoadLocation("Asia/Jakarta")
			if err != nil {
				loc = time.FixedZone("WIB", 7*3600)
			}
			now := time.Now().In(loc)

			daysIndo := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
			monthsIndo := []string{
				"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
				"Juli", "Agustus", "September", "Oktober", "November", "Desember",
			}

			dayName := daysIndo[now.Weekday()]
			monthName := monthsIndo[now.Month()]

			formatted := fmt.Sprintf("%s, %02d %s %d %02d:%02d:%02d WIB",
				dayName, now.Day(), monthName, now.Year(), now.Hour(), now.Minute(), now.Second(),
			)

			return fmt.Sprintf("Waktu sistem saat ini: %s (Tanggal: %02d, Bulan: %02d, Tahun: %d)",
				formatted, now.Day(), int(now.Month()), now.Year(),
			), nil
		},
	}
}

// NotifyTechnicianInput adalah data yang dikirimkan oleh AI untuk laporan teknisi.
type NotifyTechnicianInput struct {
	CustomerName     string `json:"customer_name" jsonschema:"description=Nama lengkap pelanggan yang melapor"`
	CustomerPhone    string `json:"customer_phone" jsonschema:"description=Nomor HP atau WhatsApp aktif pelanggan yang bisa dihubungi teknisi"`
	Address          string `json:"address" jsonschema:"description=Alamat lengkap lokasi pelanggan atau titik gangguan"`
	IssueType        string `json:"issue_type" jsonschema:"description=Kategori gangguan (contoh: Kabel Fiber Putus, Modem Rusak / Mati, Redaman Tinggi)"`
	IssueDescription string `json:"issue_description" jsonschema:"description=Detail atau keterangan kendala yang dialami"`
}

// NewNotifyTechnicianTool membuat tool LLM untuk mengirimkan notifikasi laporan gangguan langsung ke WhatsApp teknisi yang aktif di database.
func NewNotifyTechnicianTool(userRepo port.UserRepository, waGateway any, sessionID uint) llm.Tool {
	return llm.Tool{
		Name:        "notify_technician",
		Description: "Meneruskan laporan gangguan dan permintaan kunjungan teknisi lapangan langsung ke WhatsApp teknisi aktif setelah pelanggan memberikan data (Nama, Alamat, No. HP, dan Detail Kendala).",
		InputSchema: NotifyTechnicianInput{},
		Handler: func(ctx context.Context, argsJSON string) (string, error) {
			var input NotifyTechnicianInput
			if strings.TrimSpace(argsJSON) != "" {
				_ = json.Unmarshal([]byte(argsJSON), &input)
			}

			name := strings.TrimSpace(input.CustomerName)
			phone := strings.TrimSpace(input.CustomerPhone)
			addr := strings.TrimSpace(input.Address)
			issueType := strings.TrimSpace(input.IssueType)
			desc := strings.TrimSpace(input.IssueDescription)

			if name == "" || addr == "" || phone == "" {
				return "Error: Data belum lengkap. Mohon minta pelanggan melengkapi Nama, Alamat, dan Nomor HP yang dapat dihubungi.", nil
			}

			if issueType == "" {
				issueType = "Gangguan Fisik Jaringan"
			}
			if desc == "" {
				desc = "Pelanggan meminta kunjungan teknisi ke lokasi."
			}

			loc, err := time.LoadLocation("Asia/Jakarta")
			if err != nil {
				loc = time.FixedZone("WIB", 7*3600)
			}
			nowStr := time.Now().In(loc).Format("02/01/2006 15:04 WIB")

			formattedMessage := fmt.Sprintf("🚨 *LAPORAN GANGGUAN PELANGGAN* 🚨\n"+
				"━━━━━━━━━━━━━━━━━━━━━\n"+
				"👤 *Pelanggan:* %s\n"+
				"📞 *Kontak:* %s\n"+
				"📍 *Alamat:* %s\n"+
				"⚠️ *Jenis Kendala:* %s\n"+
				"📝 *Keterangan:* %s\n"+
				"⏰ *Waktu Lapor:* %s\n"+
				"━━━━━━━━━━━━━━━━━━━━━\n"+
				"_Diteruskan otomatis oleh AI CS Bot Ghaib Network._",
				name, phone, addr, issueType, desc, nowStr,
			)

			var notifiedCount int
			var notifiedNames []string

			if userRepo != nil {
				techUsers, err := userRepo.FindByRoles(ctx, []string{"teknisi", "technician"}, true)
				if err == nil {
					gw, ok := waGateway.(interface {
						SendMessage(sessionID uint, to string, content string) error
					})
					for _, tech := range techUsers {
						rawPhone := strings.TrimSpace(tech.PhoneNumber)
						if rawPhone == "" {
							continue
						}
						targetPhone := strings.TrimPrefix(rawPhone, "+")
						if strings.HasPrefix(targetPhone, "0") {
							targetPhone = "62" + targetPhone[1:]
						}
						targetJID := targetPhone + "@s.whatsapp.net"
						if strings.Contains(targetPhone, "-") || strings.Contains(targetPhone, "@g.us") {
							targetJID = targetPhone
							if !strings.HasSuffix(targetJID, "@g.us") {
								targetJID = targetJID + "@g.us"
							}
						}

						if ok && gw != nil {
							_ = gw.SendMessage(sessionID, targetJID, formattedMessage)
						}
						notifiedCount++
						displayName := tech.FullName
						if displayName == "" {
							displayName = tech.Username
						}
						notifiedNames = append(notifiedNames, displayName)
					}
				}
			}

			if notifiedCount > 0 {
				return fmt.Sprintf("Sukses! Laporan untuk %s (%s) di %s telah berhasil diteruskan ke %d teknisi aktif (%s).",
					name, phone, addr, notifiedCount, strings.Join(notifiedNames, ", "),
				), nil
			}

			return fmt.Sprintf("Sukses! Laporan untuk %s (%s) di %s telah berhasil dicatat untuk tindak lanjut tim teknisi lapangan.",
				name, phone, addr,
			), nil
		},
	}
}
