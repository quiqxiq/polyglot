package skill_test

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	"github.com/stretchr/testify/assert"
)

func TestMatchRelevantFiles(t *testing.T) {
	mockSkill := &skill.Skill{
		Slug: "ghaib-cs",
		Name: "Ghaib CS",
		Files: []skill.SkillFile{
			{FilePath: "SKILL.md", Name: "SKILL.md", Content: "# Core SOP", IsReference: false},
			{FilePath: "references/paket-dan-harga.md", Name: "paket-dan-harga.md", Content: "# Paket 10Mbps 100rb", IsReference: true},
			{FilePath: "references/troubleshooting-jaringan.md", Name: "troubleshooting-jaringan.md", Content: "# Alur LOS Merah", IsReference: true},
			{FilePath: "references/tagihan-dan-pembayaran.md", Name: "tagihan-dan-pembayaran.md", Content: "# Cara Transfer BCA", IsReference: true},
			{FilePath: "references/profil-perusahaan.md", Name: "profil-perusahaan.md", Content: "# WA Sales: 0812...", IsReference: true},
			{FilePath: "references/eskalasi-dan-faq.md", Name: "eskalasi-dan-faq.md", Content: "# Hubungi CS Manusia", IsReference: true},
		},
	}

	t.Run("General Greeting: Only SKILL.md included", func(t *testing.T) {
		matched := skillUC.MatchRelevantFiles(mockSkill, "Halo kak selamat pagi")
		assert.Len(t, matched, 1)
		assert.Equal(t, "SKILL.md", matched[0].FilePath)
	})

	t.Run("Price inquiry: SKILL.md + paket-dan-harga + profil-perusahaan", func(t *testing.T) {
		matched := skillUC.MatchRelevantFiles(mockSkill, "Berapa harga paket internet 20 mbps?")
		paths := extractFilePaths(matched)
		assert.Contains(t, paths, "SKILL.md")
		assert.Contains(t, paths, "references/paket-dan-harga.md")
		assert.Contains(t, paths, "references/profil-perusahaan.md")
		assert.NotContains(t, paths, "references/troubleshooting-jaringan.md")
		assert.NotContains(t, paths, "references/tagihan-dan-pembayaran.md")
	})

	t.Run("Troubleshooting inquiry: SKILL.md + troubleshooting-jaringan", func(t *testing.T) {
		matched := skillUC.MatchRelevantFiles(mockSkill, "Wifi saya mati lampu modem indikator LOS warna merah")
		paths := extractFilePaths(matched)
		assert.Contains(t, paths, "SKILL.md")
		assert.Contains(t, paths, "references/troubleshooting-jaringan.md")
		assert.NotContains(t, paths, "references/tagihan-dan-pembayaran.md")
	})

	t.Run("Billing inquiry: SKILL.md + tagihan-dan-pembayaran", func(t *testing.T) {
		matched := skillUC.MatchRelevantFiles(mockSkill, "Mau tanya nomor rekening untuk bayar tagihan bulanan")
		paths := extractFilePaths(matched)
		assert.Contains(t, paths, "SKILL.md")
		assert.Contains(t, paths, "references/tagihan-dan-pembayaran.md")
	})
}

func extractFilePaths(files []skill.SkillFile) []string {
	var paths []string
	for _, f := range files {
		paths = append(paths, f.FilePath)
	}
	return paths
}
