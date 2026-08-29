//go:build integration

// Test integrasi Postgres untuk UpsertMessagesBatch (multi-row
// INSERT ... ON CONFLICT DO NOTHING). Menjalankan terhadap database nyata dan
// memverifikasi: RowsAffected = jumlah baris yang benar-benar di-insert,
// idempotensi saat replay (unique session_id+wa_message_id), serta perilaku
// batch campuran (baru + lama).
//
// Tes membuat database sementara (polyglot_it_*) di server yang sama dan
// menghapusnya di akhir — tidak menyentuh data dev dan tidak bergantung pada
// skema database yang sudah ada (mis. tabel yang pernah dibuat AutoMigrate
// versi lama dengan nama kolom berbeda).
//
// Jalankan (sesuaikan DSN dengan environment lokal):
//
//	TEST_DATABASE_URL='postgres://postgres:netops@localhost:5432/netops?sslmode=disable' \
//	  go test -tags=integration ./test/integration/ -run TestUpsertMessagesBatch -v
package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	apppg "github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

// TestUpsertMessagesBatch verifies multi-row insert semantics against a real
// PostgreSQL instance: fresh inserts, idempotent replay, mixed batches, and
// final table contents.
func TestUpsertMessagesBatch(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL (atau DATABASE_URL) tidak di-set — skip integration test Postgres")
	}

	// Isolasi penuh: buat database sementara di server yang sama, jalankan
	// seluruh skenario di sana, lalu drop (WITH FORCE memutus koneksi sisa).
	tempName := "polyglot_it_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	adminDB, err := gorm.Open(gormpg.Open(replaceDBName(dsn, "postgres")), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	if err := adminDB.Exec(`CREATE DATABASE "` + tempName + `"`).Error; err != nil {
		_ = adminDB.Exec(`DROP DATABASE IF EXISTS "` + tempName + `"`).Error
		t.Fatalf("CREATE DATABASE: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Exec(`DROP DATABASE IF EXISTS "` + tempName + `" WITH (FORCE)`).Error
		if sqlDB, err := adminDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	store, err := apppg.NewStore(replaceDBName(dsn, tempName))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	db := store.DB()

	ensureWAMirrorSchema(t, db)

	// Jaring pengaman regresi kelas bug akronim GORM: kolom harus mengikuti
	// nama migrasi (jid/chat_jid/sender_jid), bukan hasil mangle GORM
	// (j_id/chat_j_id/sender_j_id). Tanpa tag `column:` eksplisit di model,
	// asersi ini akan gagal.
	assertColumn(t, db, "wa_sessions", "jid")
	assertColumn(t, db, "wa_messages", "chat_jid")
	assertColumn(t, db, "wa_messages", "sender_jid")

	sessionID := createWATestSession(t, db)

	const chatA = "6281234567890@s.whatsapp.net"
	const chatB = "6289876543210@s.whatsapp.net"
	base := time.Now().UTC().Truncate(time.Second)

	msgs := []*bot.WAMessage{
		waTestMessage(sessionID, chatA, "m1", "pesan satu", false, base),
		waTestMessage(sessionID, chatA, "m2", "pesan dua", false, base.Add(time.Minute)),
		waTestMessage(sessionID, chatB, "m3", "pesan tiga", true, base.Add(2*time.Minute)),
	}

	// 1) Batch pertama: 3 pesan baru -> 3 baris ter-insert.
	inserted, err := store.UpsertMessagesBatch(context.Background(), msgs)
	if err != nil {
		t.Fatalf("batch pertama gagal: %v", err)
	}
	if inserted != 3 {
		t.Fatalf("batch pertama: want inserted=3, got %d", inserted)
	}

	// 2) Replay identik: ON CONFLICT DO NOTHING -> 0 baris baru.
	inserted, err = store.UpsertMessagesBatch(context.Background(), msgs)
	if err != nil {
		t.Fatalf("replay gagal: %v", err)
	}
	if inserted != 0 {
		t.Fatalf("replay: want inserted=0, got %d (idempotensi unik session_id+wa_message_id rusak)", inserted)
	}

	// 3) Batch campuran: 1 baru (m4) + 2 lama (m1, m2) -> hanya 1 yang masuk.
	mixed := []*bot.WAMessage{
		waTestMessage(sessionID, chatA, "m4", "pesan baru", false, base.Add(3*time.Minute)),
		msgs[0],
		msgs[1],
	}
	inserted, err = store.UpsertMessagesBatch(context.Background(), mixed)
	if err != nil {
		t.Fatalf("batch campuran gagal: %v", err)
	}
	if inserted != 1 {
		t.Fatalf("batch campuran: want inserted=1, got %d", inserted)
	}

	// 4) Slice kosong -> (0, nil) tanpa query.
	inserted, err = store.UpsertMessagesBatch(context.Background(), nil)
	if err != nil || inserted != 0 {
		t.Fatalf("batch kosong: want (0, nil), got (%d, %v)", inserted, err)
	}

	// 5) Verifikasi isi tabel lewat query mentah + API publik.
	var total int64
	if err := db.Raw("SELECT count(*) FROM wa_messages WHERE session_id = ?", sessionID).
		Scan(&total).Error; err != nil {
		t.Fatalf("count pesan gagal: %v", err)
	}
	if total != 4 {
		t.Fatalf("total pesan: want 4, got %d", total)
	}

	msgsA, err := store.ListChatMessages(context.Background(), sessionID, chatA, 50, 0)
	if err != nil {
		t.Fatalf("ListChatMessages: %v", err)
	}
	if len(msgsA) != 3 {
		t.Fatalf("chatA: want 3 pesan, got %d", len(msgsA))
	}
	wantContents := []string{"pesan satu", "pesan dua", "pesan baru"}
	for i, want := range wantContents {
		if msgsA[i].Content != want {
			t.Fatalf("chatA urutan[%d]: want %q, got %q (order timestamp ASC rusak)", i, want, msgsA[i].Content)
		}
	}
}

// waTestMessage builds a minimal WAMessage mirror row for the batch test.
func waTestMessage(sessionID uint, chatJID, waID, content string, fromMe bool, ts time.Time) *bot.WAMessage {
	return &bot.WAMessage{
		SessionID:   sessionID,
		ChatJID:     chatJID,
		WAMessageID: waID,
		SenderJID:   chatJID,
		SenderName:  "Tester",
		Content:     content,
		MediaType:   "text",
		IsFromMe:    fromMe,
		Timestamp:   ts,
	}
}

// ensureWAMirrorSchema creates wa_sessions/wa_chats/wa_messages when missing
// (idempotent), mirroring the DDL of migrations 000002/000004. Ketiga tabel
// biasanya sudah dibuat oleh AutoMigrate NewStore (model kini selaras dengan
// migrasi: tag `column:` untuk ChatJID/SenderJID dan TableName WASessionModel
// -> wa_sessions); DDL mentah di sini hanyalah jaring pengaman untuk skema
// yang dibuat tanpa AutoMigrate.
func ensureWAMirrorSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS wa_sessions (
			id SERIAL PRIMARY KEY,
			device_name VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50),
			jid VARCHAR(255),
			status VARCHAR(50) DEFAULT 'offline',
			is_bot_enabled BOOLEAN DEFAULT TRUE,
			connected_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS wa_chats (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES wa_sessions(id) ON DELETE CASCADE,
			chat_jid VARCHAR(255) NOT NULL,
			display_name VARCHAR(255),
			is_group BOOLEAN NOT NULL DEFAULT FALSE,
			last_message_id VARCHAR(255),
			last_message_preview TEXT,
			last_message_time TIMESTAMP WITH TIME ZONE,
			unread_count INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_wa_chats_session_jid UNIQUE (session_id, chat_jid)
		)`,
		`CREATE TABLE IF NOT EXISTS wa_messages (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES wa_sessions(id) ON DELETE CASCADE,
			chat_jid VARCHAR(255) NOT NULL,
			wa_message_id VARCHAR(255) NOT NULL,
			sender_jid VARCHAR(255),
			sender_name VARCHAR(255),
			content TEXT,
			media_type VARCHAR(50) NOT NULL DEFAULT 'text',
			is_from_me BOOLEAN NOT NULL DEFAULT FALSE,
			is_read BOOLEAN NOT NULL DEFAULT FALSE,
			timestamp TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_wa_messages_session_wa_id UNIQUE (session_id, wa_message_id)
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("ensure schema: %v", err)
		}
	}
}

// assertColumn fails the test when a table is missing the given column (or has
// it under a GORM-mangled variant of the same acronym).
func assertColumn(t *testing.T, db *gorm.DB, table, column string) {
	t.Helper()
	var count int64
	if err := db.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_name = ? AND column_name = ?`, table, column).Scan(&count).Error; err != nil {
		t.Fatalf("cek kolom %s.%s gagal: %v", table, column, err)
	}
	if count == 0 {
		t.Fatalf("kolom %s.%s tidak ada — model GORM divergen dari migrasi (kemungkinan akronim ter-mangle)", table, column)
	}
}

// createWATestSession inserts a throwaway wa_sessions row and returns its id.
func createWATestSession(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	var id uint
	name := fmt.Sprintf("it-batch-%d", time.Now().UnixNano())
	if err := db.Raw("INSERT INTO wa_sessions (device_name) VALUES (?) RETURNING id", name).
		Scan(&id).Error; err != nil {
		t.Fatalf("insert session: %v", err)
	}
	return id
}

// replaceDBName swaps the database path segment of a postgres URL-form DSN
// (mis. `postgres://user:pass@host:5432/netops?sslmode=disable`). DSN bentuk
// key=value (tanpa `/`) tidak didukung dan dikembalikan apa adanya — gunakan
// bentuk URL pada TEST_DATABASE_URL/DATABASE_URL.
func replaceDBName(dsn, dbName string) string {
	base, query, _ := strings.Cut(dsn, "?")
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[:i+1] + dbName
	}
	if query != "" {
		return base + "?" + query
	}
	return base
}
