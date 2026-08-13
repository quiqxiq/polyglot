package whatsapp

import (
	"context"
	"log"
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
		log.Printf("[JID] Cannot resolve LID %s: client not available", jid.String())
		return jid
	}
	pn, err := client.Store.LIDs.GetPNForLID(ctx, jid)
	if err != nil {
		log.Printf("[JID] Failed to resolve LID %s: %v", jid.String(), err)
		return jid
	}
	if pn.IsEmpty() {
		return jid
	}
	return pn
}
