package whatsapp

import (
	"context"
	"github.com/quixiq/polyglot/pkg/logger"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// isSystemBroadcastJID melaporkan apakah jid adalah sistem broadcast WhatsApp
// yang TIDAK boleh masuk pipeline chat: status@broadcast (feed story per-akun)
// dan 0@s.whatsapp.net (akun resmi layanan WhatsApp). Match dilakukan eksak
// (bukan suffix) agar JID seperti "status@s.whatsapp.net" tidak ikut tertangkap.
func isSystemBroadcastJID(jid string) bool {
	return jid == "status@broadcast" || jid == "0@s.whatsapp.net"
}

// isNewsletterJID melaporkan apakah jid adalah channel WhatsApp (newsletter).
// Channel adalah broadcast feed, bukan percakapan — local part-nya berupa
// 18-digit channel ID, bukan nomor HP, sehingga tidak bisa diperlakukan sebagai
// chat 1:1 maupun grup biasa.
func isNewsletterJID(jid string) bool {
	return strings.HasSuffix(jid, "@newsletter")
}

// isSkippedJID adalah satu titik panggil gabungan: mengembalikan true untuk
// setiap JID yang harus DIABAIKAN sepenuhnya — tidak ditulis ke mirror DB,
// tidak diteruskan ke bot engine, tidak ditampilkan di ListChats.
func isSkippedJID(jid string) bool {
	return isSystemBroadcastJID(jid) || isNewsletterJID(jid)
}

// resolveChatDisplayName menentukan nama tampil sebuah chat:
//   - Grup: nama grup dari history sync (fallback) — nama pengirim tidak pernah
//     dipakai sebagai nama chat.
//   - 1:1: prioritas nama tersimpan pengguna (FullName/FirstName dari contact
//     store whatsmeow, yang diisi dari buku alamat HP) → push name → fallback
//     (push name pesan) → nomor HP asli (setelah normalisasi LID→PN).
//
// Ini meniru perilaku GetChatNameWithPushName pada referensi
// go-whatsapp-web-multidevice, dan menyelesaikan dua masalah sekaligus:
// nama kontak tidak muncul dan chat @lid menampilkan angka LID.
func (c *Client) resolveChatDisplayName(ctx context.Context, jid types.JID, fallback string) string {
	if jid.Server == types.GroupServer {
		return fallback
	}
	pnJID := normalizeJIDFromLID(ctx, jid, c.waClient)
	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.Contacts != nil {
		if info, err := c.waClient.Store.Contacts.GetContact(ctx, pnJID); err == nil {
			if info.FullName != "" {
				return info.FullName
			}
			if info.FirstName != "" {
				return info.FirstName
			}
			if info.PushName != "" {
				return info.PushName
			}
		}
	}
	if fallback != "" {
		return fallback
	}
	return pnJID.ToNonAD().User
}

// normalizeJIDFromLID mengkonversi JID @lid ke JID @s.whatsapp.net yang sesuai
// (nomor HP nyata). Dibutuhkan karena WhatsApp mulai mengirimkan JID berformat
// @lid untuk sebagian kontak (privacy-preserving linked IDs). Tanpa normalisasi,
// sender tampil sebagai "12345@lid" bukan nomor HP.
//
// Mengembalikan jid asli tanpa modifikasi bila:
//   - jid bukan @lid
//   - client/store tidak tersedia (belum connect)
//   - lookup gagal (LID belum diketahui store)
func normalizeJIDFromLID(ctx context.Context, jid types.JID, client *whatsmeow.Client) types.JID {
	if jid.Server != "lid" {
		return jid
	}
	if client == nil || client.Store == nil || client.Store.LIDs == nil {
		logger.WithComponent("JIDFilter").Warn("cannot resolve LID; client unavailable")
		return jid
	}
	pn, err := client.Store.LIDs.GetPNForLID(ctx, jid)
	if err != nil {
		logger.WithComponent("JIDFilter").WithError(err).Warn("failed to resolve LID")
		return jid
	}
	if pn.IsEmpty() {
		return jid
	}
	return pn
}
