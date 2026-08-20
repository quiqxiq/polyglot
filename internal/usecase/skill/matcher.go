package skill

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// referenceTopicRules memetakan kata kunci pencarian ke slug atau nama berkas referensi.
var referenceTopicRules = []struct {
	FileKeyword string   // Substring pencocokan nama file (misal "paket", "troubleshoot", "tagihan")
	Triggers    []string // Kata kunci dalam pesan user
}{
	{
		FileKeyword: "paket",
		Triggers: []string{
			"paket", "harga", "biaya", "speed", "mbps", "pasang", "promo",
			"langganan", "daftar", "tarif", "beli", "upgrade", "downgrade",
			"berhenti", "relokasi", "unlimited", "fup", "kuota",
		},
	},
	{
		FileKeyword: "troubleshoot",
		Triggers: []string{
			"gangguan", "mati", "los", "merah", "lemot", "lambat", "rusak",
			"koneksi down", "koneksi lambat", "koneksi putus", "modem", "router",
			"ont", "kabel", "restart", "sinyal hilang", "putus", "pon", "bisa connect",
			"tidak ada internet", "tanda seru", "lampu", "trouble", "kendala",
		},
	},
	{
		FileKeyword: "tagihan",
		Triggers: []string{
			"tagihan", "bayar", "transfer", "rekening", "bca", "bri", "mandiri",
			"dana", "tempo", "jatuh tempo", "invois", "invoice", "struk",
			"isolir", "reaktivasi", "biaya bulanan", "pembayaran", "bukti bayar",
		},
	},
	{
		FileKeyword: "profil",
		Triggers: []string{
			"profil", "alamat", "kantor", "kontak", "email", "telepon", "coverage",
			"lokasi", "area", "jangkauan", "sales", "cs", "nomor admin",
			"siapa ghaib", "perusahaan", "legalitas", "wa admin",
		},
	},
	{
		FileKeyword: "eskalasi",
		Triggers: []string{
			"teknisi", "komplain", "admin", "bantuan", "manusia", "orang",
			"faq", "lapor", "marah", "kecewa", "bicara langsung", "hubungi staf",
			"kunjungan", "survei",
		},
	},
}

// MatchRelevantFiles menyaring berkas-berkas di dalam Skill agar hanya berkas yang relevan dengan pertanyaan user yang disertakan.
// Berkas utama (SKILL.md) selalu disertakan sebagai pedoman inti.
func MatchRelevantFiles(sk *skill.Skill, contextText string) []skill.SkillFile {
	if sk == nil || len(sk.Files) == 0 {
		return nil
	}

	ctxLower := strings.ToLower(strings.TrimSpace(contextText))

	var matchedFiles []skill.SkillFile
	var referenceFiles []skill.SkillFile

	// 1. Pisahkan SKILL.md (berkas utama) dari berkas referensi
	for _, f := range sk.Files {
		filePathLower := strings.ToLower(f.FilePath)
		if filePathLower == "skill.md" || strings.HasSuffix(filePathLower, "/skill.md") || !f.IsReference {
			matchedFiles = append(matchedFiles, f)
		} else {
			referenceFiles = append(referenceFiles, f)
		}
	}

	// 2. Jika contextText kosong, kembalikan hanya berkas utama tanpa referensi berat
	if ctxLower == "" {
		return matchedFiles
	}

	// 3. Cari berkas referensi yang relevan dengan kata kunci
	selectedRefMap := make(map[string]bool)

	for _, ref := range referenceFiles {
		refPathLower := strings.ToLower(ref.FilePath)
		refNameLower := strings.ToLower(ref.Name)

		for _, rule := range referenceTopicRules {
			if strings.Contains(refPathLower, rule.FileKeyword) || strings.Contains(refNameLower, rule.FileKeyword) {
				for _, kw := range rule.Triggers {
					if strings.Contains(ctxLower, kw) {
						if !selectedRefMap[ref.FilePath] {
							selectedRefMap[ref.FilePath] = true
							matchedFiles = append(matchedFiles, ref)
						}
						break
					}
				}
			}
		}
	}

	// 4. Jika user bertanya seputar kontak/paket tapi profil-perusahaan belum terikut, tambahkan profil jika ada kontak
	if selectedRefMap["references/paket-dan-harga.md"] || selectedRefMap["references/eskalasi-dan-faq.md"] {
		for _, ref := range referenceFiles {
			if strings.Contains(strings.ToLower(ref.FilePath), "profil") && !selectedRefMap[ref.FilePath] {
				selectedRefMap[ref.FilePath] = true
				matchedFiles = append(matchedFiles, ref)
			}
		}
	}

	return matchedFiles
}
