# 0002 — port.DeviceDriver Tanpa Tipe Session Terpisah

## Status
Diterima

## Konteks
Draf awal (`Polyglot-Architecture.md` §5.3) mendefinisikan
`Connect(ctx, target) -> Session` terpisah dari
`Execute(ctx, sess, cmd)`, plus method `Stream` langsung ada di interface.
Ini menambah satu tipe (Session) yang harus diimplementasikan tiap vendor
tanpa manfaat langsung selama connection pooling belum genuinely dibangun,
dan menambah method `Stream` yang belum akan dipakai sampai Fase 7.

## Keputusan
`NewDriver(ctx, target)` per vendor connect langsung dan mengembalikan
`*Driver` yang sudah terhubung. `Execute`/`Classify`/`Translate`/`Close`
adalah method pada `*Driver` itu. Tidak ada tipe `Session` terpisah pada
level koneksi driver. `Translate(op) (Command, error)` ditambahkan
(tidak ada di sketsa v5.3) supaya usecase punya cara vendor-agnostic untuk
menerjemahkan operasi abstrak (`get_status`, `reboot`) tanpa tahu sintaks
native vendor. `Stream` DITUNDA, tidak ditambahkan sampai Fase 7 benar-benar
dikerjakan.

## Konsekuensi
- `internal/registry` bertanggung jawab menyimpan/reuse `*Driver` per device
  (bukan per Session) untuk menghindari reconnect storm.
- `internal/domain/session/` TETAP ada sesuai struktur definitif di
  `CLAUDE.md` §1.1, tapi cakupannya bukan connection handle DeviceDriver —
  itu entity riwayat/audit sesi (tabel `sessions` di
  `Polyglot-Architecture.md` §7.2). Lihat komentar di
  `internal/domain/session/session.go`.
- **DEVIASI TERCATAT dari `Polyglot-Architecture.md` §5.3**: dokumen tersebut
  masih menampilkan sketsa `Connect/Session/Stream` sebagai contoh kode.
  ADR ini secara eksplisit menggantikannya untuk implementasi nyata. Kalau
  `Polyglot-Architecture.md` direvisi berikutnya, §5.3 sebaiknya diperbarui
  agar konsisten dengan ADR ini — sampai saat itu, ADR ini yang berlaku
  untuk kode (selaras `CLAUDE.md` §5.2: "`CLAUDE.md` adalah satu-satunya
  sumber kebenaran struktur folder/file", dan interface kontrak port/
  mengikuti kode yang sudah dibangun, bukan sketsa awal di dokumen
  arsitektur). Kalau nanti benar-benar perlu multiplexing banyak sesi per
  device dalam satu Driver, revisit ADR ini — jangan diam-diam menambah
  Session tanpa ADR baru.
