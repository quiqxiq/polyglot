Analisis: Chat History, Live Chat & Live Chat Status (vs referensi  go-whatsapp-web-multidevice )

1. Bukti dari sistem Anda (kenapa yang tampil story, bukan history chat)

Query langsung ke DB dan log server menunjukkan kondisi nyata:

 wa_chats  (session 1) — hanya 3 baris:

┌───────────────────────────────┬─────────────────────────────────────────────────┬──────────────────────────┐
│ chat_jid                      │ display_name                                    │ isi                      │
├───────────────────────────────┼─────────────────────────────────────────────────┼──────────────────────────┤
│ status@broadcast              │ "Babe_muda" (nama fallback dari pushname story) │ 19 pesan story           │
│ 120363152349863327@newsletter │ —                                               │ 1 pesan channel Newsweek │
│ 120363202150126650@newsletter │ —                                               │ 1 pesan channel Kompas   │
└───────────────────────────────┴─────────────────────────────────────────────────┴──────────────────────────┘

Log server tidak pernah mencatat  History sync: mirroring N conversations  — artinya tidak ada satupun  events.HistorySync  yang membawa percakapan. Yang masuk hanyalah live  events.Message  dari  status@broadcast  (story) dan  @newsletter  (channel), lalu  persistMirrorMessage  menuliskannya ke  wa_chats / wa_messages  seperti chat biasa.

Kesimpulan: chat 1:1 & grup nyata memang belum pernah masuk — bukan salah render, melainkan datanya tidak ada. Yang memenuhi daftar adalah story + channel karena itulah satu-satunya "pesan" yang tiba.

2. Bagaimana referensi mengimplementasikan ketiga hal itu

a) Chat history (sinkronisasi percakapan)

- Referensi menulis setiap blob history sync ke file JSON dulu ( history-{id}-{jid}-{syncType}.json ), lalu  processHistorySync  memproses ke DB. Sync type yang diproses ke DB:  INITIAL_BOOTSTRAP  +  RECENT  (conversations) dan  PUSH_NAME  (nama kontak) — sama dengan polyglot.
- Sumber data utama: WhatsApp mengirim  HistorySyncNotification  → whatsmeow auto-download ( handleHistorySyncNotificationLoop  →  DownloadHistorySync ) → fire  events.HistorySync . Jumlah chat yang dikirim ditentukan HP — saat pairing harus mencentang "Tampilkan chat terbaru / Show recent chats". Tanpa itu, HP hanya mengirim bootstrap minimal (status + newsletter), persis gejala yang Anda lihat.
- Setelah pair sukses, referensi pada  events.AppStateSyncComplete  mengirim presence ( SendPresence ) — ini menandakan device aktif sehingga HP mulai mengalirkan sinkronisasi.
- Penyimpanan: chat  StoreChat  + pesan  StoreMessagesBatch  (transaksi), hanya chat yang punya pesan ( len(messageBatch) > 0 ).
- Penanganan JID khusus (kritis):
-  status@broadcast  → selalu diberi nama "Status" ( GetChatNameWithPushName ), tidak pernah pakai nama fallback.
-  IsSystemBroadcastJID()  =  status@broadcast  &  0@s.whatsapp.net  → diblokir dari alur customer (Chatwoot), dan  @newsletter  di-skip ( IsNewsletterJID ).
- LID → nomor HP dinormalisasi ( NormalizeJIDFromLID  via  client.Store.LIDs.GetPNForLID ) untuk chat & sender, sehingga tidak muncul JID  @lid  di UI.

b) Live chat

-  events.Message  →  handleMessage  →  CreateMessage  (simpan DB) + payload webhook + broadcast websocket ke UI → daftar chat & pesan ter-update instan.
- Polyglot sudah punya padanan ini:  persistMirrorMessage  →  chat_update  SSE → invalidate query chats/messages (lihat  use-whatsapp-sse.ts ). ✅ Alurnya benar — masalahnya bukan mekanisme live-nya, tapi story/newsletter ikut lewat jalur ini.

c) Live chat status (typing/recording)

- Referensi:  events.ChatPresence  →  handleChatPresence  → deteksi  types.ChatPresenceComposing  (+  ChatPresenceMediaAudio  = sedang merekam) vs  paused  → payload  {event:"chat_presence", from, chat_id, state, media, is_group}  → broadcast websocket + webhook.
- Juga  events.Presence  (online/offline) dan  events.Receipt  (delivered/read) → status kiriman.
- Polyglot: belum ada sama sekali —  handleEvent  hanya menangani Message/Connected/Disconnected/HistorySync/LoggedOut. Tidak ada  case *events.ChatPresence , tidak ada event SSE  chat_presence , tidak ada indikator "mengetik…" di frontend.

3. Akar masalah polyglot (dari yang paling berdampak)

┌─────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┬───────────────────────────────────────────────────┐
│ #   │ Masalah                                                                                                                                                   │ Lokasi                                            │
├─────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
│ A   │ History chat nyata tidak pernah masuk — tanpa centang "Tampilkan chat terbaru" saat scan QR, HP hanya kirim bootstrap minimal; events.HistorySync berisi  │ proses pairing (manual)                           │
│     │ conversations tidak pernah tiba (nol log).                                                                                                                │                                                   │
│ B   │ Story (status@broadcast) & channel (@newsletter) di-mirror & dirender sebagai chat biasa, bahkan memicu engine bot (handleIncomingMessage → balasan ke    │ client.go: handleEvent/persistMirrorMessage/handl │
│     │ story!) dan menciptakan Conversation palsu. Tidak ada filter IsSystemBroadcastJID/newsletter, tidak ada section "Status" terpisah, display name pakai     │ eIncomingMessage, history_sync.go,                │
│     │ fallback pushname ("Babe_muda") bukan "Status".                                                                                                           │ chat_repository.go: ListChats                     │
│ C   │ LID tidak dinormalisasi → sender tampil 105952632143924@lid (story) / nomor @lid aneh, bukan nomor HP. Referensi normalisasi via GetPNForLID.             │ history_sync.go, client.go                        │
│ D   │ handleHistorySync membuang FULL/ON_DEMAND di cabang default — padahal tipe itu membawa conversations lengkap saat user membuka chat di HP (referensi      │ history_sync.go                                   │
│     │ menulisnya ke file walau tidak ke DB; di polyglot lebih baik diproses).                                                                                   │                                                   │
└─────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┴───────────────────────────────────────────────────┘

4. Rencana perbaikan (berjenjang)

Fase 1 — Berhenti menampilkan story sebagai chat & hentikan bot menjawab story:
-  extractMessageBody / persistMirrorMessage / handleIncomingMessage / processHistoryConversations : skip JID  status@broadcast ,  0@s.whatsapp.net ,  @newsletter  ( types.NewsletterServer  / suffix  newsletter ).
-  ListChats  (SQL) tambah  WHERE chat_jid NOT IN (...)  untuk sistem JID.
- Bot engine: blokir callback untuk sistem JID.

Fase 2 — Label & tampilkan Status dengan benar (opsional, mengikuti referensi):
- Simpan  status@broadcast  dengan display name tetap "Status" +  is_group=false , dan di frontend render sebagai section/entry terpisah (mirip WhatsApp Web), bukan campur di daftar chat.

Fase 3 — Live chat status (typing/recording):
- Backend:  case *events.ChatPresence:  di  handleEvent  → callback → broadcast SSE  chat_presence   {session_id, chat_jid, sender_jid, state: composing|paused, media} .
- Frontend: konsumen SSE baru + indikator "mengetik…" di header chat + subscribe saat chat aktif.

Fase 4 — Kualitas data (menyamai referensi):
- Normalisasi LID→PN untuk sender (pakai  client.Store.LIDs.GetPNForLID ).
-  handleHistorySync : proses juga  FULL / ON_DEMAND  (bukan buang).
- (Verifikasi manual) Saat re-pair ulang, centang "Tampilkan chat terbaru" di dialog HP →  RECENT  sync masuk → seluruh chat 1:1 & grup muncul.

Catatan penting: Fase 1-4 tidak bisa memunculkan history chat yang tidak pernah dikirim HP. Untuk melihat chat nyata, wajib re-pair dengan centang "Tampilkan chat terbaru" (atau setelan HP: WhatsApp → Perangkat tertaut → pilih perangkat → "Tampilkan chat terbaru"). Itu sebabnya referensi berfungsi — ia dipakai sebagai aplikasi WhatsApp Web aktif, bukan mirror pasif.