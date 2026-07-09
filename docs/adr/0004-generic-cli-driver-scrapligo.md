# 0004 — genericssh & generictelnet Berbagi Satu Mesin scrapligo

## Status
Diterima

## Konteks

Pertanyaan yang memicu ADR ini: apakah SSH dan Telnet, sebagai transport
untuk CLI vendor jaringan, benar-benar "konsep yang sama, cuma command yang
beda" — dan kalau iya, apakah `genericssh` dan `generictelnet` bisa berbagi
satu implementasi.

Jawabannya, dikonfirmasi langsung dari source `github.com/scrapli/scrapligo`
v1.3.3 (bukan tebakan): **ya, sebagian besar benar, dengan satu nuansa
penting.**

- scrapligo sudah punya transport Telnet asli (`transport/telnet.go`),
  termasuk negosiasi IAC (`do`/`dont`/`will`/`wont`) — bukan sesuatu yang
  perlu ditulis dari nol.
- Mesin "generic"-nya (deteksi prompt, disable paging, privilege
  escalation, deteksi command gagal) di scrapligo memang transport-agnostic
  — didorong oleh definisi platform (YAML), dipilih transportnya lewat
  `options.WithTransportType("standard"|"telnet"|"system")`. Definisi
  platform yang sama bentuknya persis sama dipakai baik untuk SSH maupun
  Telnet.
- **Nuansa**: "concept sama, command beda" itu benar untuk *transport*
  (dial + baca/tulis byte), tapi deteksi prompt/paging/privilege-escalation
  BUKAN "command" — itu bagian dari mesin generic yang sama, didorong oleh
  **data** (YAML), bukan oleh kode Go yang beda per vendor. Jadi
  kesimpulannya lebih tepat: "satu mesin generic, dua transport, N vendor
  — semuanya beda di DATA (platform YAML + katalog command), bukan di kode
  Go."

## Keputusan

1. **Satu paket internal bersama, `internal/driver/genericcli`**, dipakai
   oleh `internal/driver/genericssh` dan `internal/driver/generictelnet`.
   Keduanya cuma beda `transportType` ("standard" vs "telnet") dan
   `defaultPort` (22 vs 23) yang dioper ke `genericcli.NewSession`. Semua
   logic Execute/reconnect/serialisasi wire ada di `genericcli` sekali
   saja.
   - DEVIASI tercatat dari pola "satu paket = satu vendor" yang berlaku
     untuk driver konkret (mikrotik, cisco, dst): `genericssh` dan
     `generictelnet` masing-masing tetap py package sendiri (dengan
     `var _ port.DeviceDriver = (*Driver)(nil)` sendiri, sesuai CLAUDE.md
     §1.2), tapi keduanya cuma wrapper tipis di atas `genericcli.Session`.

2. **Transport SSH: `"standard"` (crypto/ssh murni), BUKAN `"system"`**
   (default scrapligo, membungkus binary `/bin/ssh`). `"system"` akan
   merusak model deployment single-binary proyek ini (butuh `ssh` binary
   terpasang di container/host). Ini keputusan eksplisit, bukan default
   yang dibiarkan begitu saja.

3. **Vendor sepenuhnya jadi DATA, bukan kode**, lewat dua parameter yang
   dioper ke `NewDriver`:
   - `platformDef any` — nama platform bawaan scrapligo (mis.
     `"cisco_iosxe"`), path file YAML custom (`internal/platformdef/`),
     atau bytes YAML/JSON mentah. Diteruskan apa adanya ke
     `platform.NewPlatform`.
   - `genericcli.Catalog` — pengganti `commands.go` untuk driver konkret:
     `DestructivePrefixes`, `Operations` (peta `command.Operation` →
     `command.Command`), dan `FailedWhenContains`.

   Konsekuensi: `genericssh`/`generictelnet` **tidak punya `commands.go`**
   — DEVIASI tercatat dari pola driver konkret lain, karena memang tidak
   ada katalog tetap untuk disimpan di situ.

4. **`resolveDriver` mencoba `GetNetworkDriver()` dulu, baru
   `GetGenericDriver()`** — ditemukan langsung dari source
   `platform/definition.go`: `Platform.GetNetworkDriver()` ERROR kalau
   YAML-nya mendeklarasikan `driver-type: generic`, dan sebaliknya. Semua
   platform bawaan scrapligo DAN template resmi
   (`assets/platforms/example.yaml`) memakai `driver-type: network`, jadi
   itu dicoba lebih dulu; `generic` jadi fallback. `genericcli.Session`
   bekerja dengan keduanya lewat interface `cliDriver` yang sama, tidak
   pernah hardcode salah satunya.

5. **Model resiliensi lebih SEDERHANA dibanding mikrotik**: satu kali
   retry-setelah-reconnect di `Execute`, bukan supervisor goroutine
   background seperti `mikrotik.Driver`. Alasan: go-routeros (dipakai
   mikrotik) punya channel error dari `Async()` yang bisa "ditunggu" tanpa
   polling; scrapligo tidak punya mekanisme setara — supervisor background
   di sini hanya bisa dibuat lewat polling, dan proyek ini sudah menetapkan
   "tidak ada polling" sebagai prinsip. Retry-sinkron-sekali adalah
   alternatif yang jujur, bukan penyederhanaan yang didiamkan.

6. **`context.Context` yang dioper ke pustaka scrapligo TIDAK mengandung
   risiko yang sama seperti go-routeros** (lihat ADR 0003): `Open()` dan
   `SendCommand()` scrapligo v1.3.3 sama sekali tidak menerima parameter
   `context.Context` — jadi tidak ada jalan bagi `ctx` milik caller untuk
   "meracuni" koneksi yang sedang berjalan. `ctx` di `Session.Execute`
   murni dipakai driver ini sendiri (goroutine + `select`) untuk
   memutuskan berapa lama menunggu — pola yang sama persis dengan mikrotik,
   tapi karena alasan yang lebih sederhana (pustakanya memang tidak
   mendukung `ctx`, bukan karena mendukungnya secara berbahaya).

7. **`internal/driver/zteolt` SENGAJA TIDAK disentuh/dipindah ke
   `generictelnet`.** `TECH-STACK-DAN-PERSIAPAN.md` §8 sudah punya
   keputusan terdokumentasi: ZTE OLT tidak punya platform definition
   scrapligo dan pakai klien Telnet mentah sebagai "satu-satunya jalan
   realistis". Menemukan bahwa scrapligo PUNYA transport Telnet generik
   tidak otomatis membatalkan keputusan itu — kalau ZTE OLT ternyata cocok
   dipindah ke `generictelnet` + YAML custom, itu keputusan baru yang butuh
   ADR-nya sendiri dan validasi nyata ke device ZTE, bukan asumsi diam-diam
   di ADR ini.

## Konsekuensi

- Menambah dukungan vendor baru lewat jalur generic ini = menulis satu
  file YAML (`internal/platformdef/`) + satu `genericcli.Catalog` di sisi
  pemanggil (`internal/registry`, saat itu dibangun) — TIDAK menulis paket
  Go baru. Kalau vendor itu ternyata butuh logic khusus di luar yang bisa
  diekspresikan lewat YAML + Catalog (mis. dua koneksi seperti mikrotik,
  atau privilege model yang aneh), itu sinyal untuk "naik kelas" jadi
  paket `internal/driver/<vendor>` sendiri — bukan memaksakan
  `genericssh`/`generictelnet` menangani kasus itu.
- `NewDriver` di `genericssh`/`generictelnet` punya 4 parameter
  (`ctx, target, platformDef, catalog`), BEDA dari pola 2-parameter
  (`ctx, target`) di semua driver konkret lain. Siapa pun yang membangun
  `internal/registry` nanti perlu tahu ini — bukan bug, tapi juga bukan
  sesuatu yang boleh "diseragamkan" tanpa mikir ulang constructor driver
  konkret lain.
- Butuh 5 modul Go baru: `github.com/scrapli/scrapligo`,
  `golang.org/x/crypto`, `golang.org/x/net`, `golang.org/x/sys`,
  `golang.org/x/text` (dependensi transitif scrapligo). Jalankan
  `go get github.com/scrapli/scrapligo@v1.3.3` di lingkungan dengan akses
  internet normal — `golang.org/x/*` akan ter-resolve langsung dari
  `golang.org`/`proxy.golang.org`, tidak perlu redirect ke mirror GitHub
  manapun (itu murni workaround sandbox pengembangan, bukan kebutuhan
  project).

## Addendum — `internal/platformdef/mikrotik_routeros.yaml`

Ditambahkan sebagai contoh nyata pertama platformdef custom (semua vendor
lain yang sudah disebut di proyek ini — Cisco, Huawei VRP — ternyata
punya platform BAWAAN scrapligo, jadi tidak butuh file YAML custom sama
sekali; lihat daftar `assets/platforms/` di source scrapligo). Dua
keputusan tambahan yang muncul saat menyusunnya:

1. **`genericcli.Catalog` bertambah field `ReadDelay time.Duration`.**
   RouterOS punya isu nyata dan terdokumentasi (bukan hipotetis): meng-echo
   prompt/command DUA KALI lewat SSH, yang pernah bikin scrapligo berhenti
   menunggu sebelum echo kedua datang (lihat
   `github.com/scrapli/scrapligo` issue #95). Solusinya
   (`options.WithReadDelay`) adalah opsi level-koneksi yang tidak bisa
   diekspresikan lewat YAML — jadi ditambahkan ke `Catalog`, bukan jadi
   parameter ke-5 di `NewSession`/`NewDriver` (lihat CLAUDE.md §3.4: lebih
   dari 4 parameter wajib struct — `Catalog` sudah jadi "struct pengganti"
   itu).

2. **Suffix `+cet512w` di username WAJIB diset pemanggil, bukan lewat
   YAML.** RouterOS mendukung parameter login (disable warna, dumb
   terminal, disable auto-detect, lebar terminal) lewat suffix di
   *username itu sendiri* — bukan lewat command apa pun setelah sesi
   terbuka (dikonfirmasi dari dokumentasi resmi MikroTik). Skema YAML
   scrapligo tidak punya field untuk memodifikasi username, jadi ini
   didokumentasikan sebagai kewajiban di sisi pemanggil
   (`device.Target.Username`), dicontohkan konkret di
   `test/integration/mikrotik_ssh_test.go`.

**Yang SUDAH divalidasi tanpa device fisik** (dari sandbox pengembangan):
skema YAML valid dan berhasil di-parse `platform.NewPlatform` +
`GetNetworkDriver`; regex prompt teruji cocok untuk prompt
normal/identity-berspasi/safe-mode/prompt-dengan-path-submenu, dan tidak
cocok untuk baris output biasa (tidak ada false positive di kasus yang
diuji).

**Yang BELUM divalidasi**: perilaku nyata ke device RouterOS fisik —
apakah `ReadDelay: 100ms` cukup, apakah `failed-when-contains` sudah
lengkap, apakah login dengan suffix `+cet512w` benar-benar menghilangkan
gejala echo ganda di issue #95. `test/integration/mikrotik_ssh_test.go`
disiapkan persis untuk validasi ini — jalankan sebelum memakainya di luar
lab.

