# 0006 — Klasifikasi Perintah Fail-Safe (Destruktif secara Default)

## Status
Diterima

## Konteks

Setiap `port.DeviceDriver` mengklasifikasikan perintah lewat
`Classify(cmd) command.Class` menjadi `ClassReadOnly` (auto-approve) atau
`ClassDestructive` (butuh HITL approval — lihat `internal/domain/command/policy.go`).
Gate ini adalah satu-satunya pembatas antara "AI/klien meminta sesuatu" dan
"perangkat benar-benar berubah" (`usecase/network.ExecuteCommand`).

Sampai ADR ini, driver mikrotik (dan pola yang sama di cisco/netconf/zteolt/
huaweiolt/genieacs) memakai posture **fail-OPEN**: sebuah daftar path yang
diketahui destruktif (`destructivePaths`), dan **apa pun di luar daftar itu
default ke `ClassReadOnly`** → auto-approve.

Selama driver hanya punya operasi baca (`/system/resource/print`) + beberapa
aksi destruktif terkenal (`/system/reboot`), lubangnya tidak terlihat. Begitu
provisioning WRITE masuk (`/ppp/secret/add`, lalu `/ppp/profile/add`,
`/ppp/profile/set`, `/ppp/active/remove` — lihat langkah 2 pada
`internal/domain/provision`), posture ini menjadi berbahaya:

- Sebuah path yang mengubah state tapi **lupa** didaftarkan ke `destructivePaths`
  akan diam-diam auto-approve. Kegagalan menambah satu baris map = eksekusi
  perubahan perangkat tanpa persetujuan.
- Beban pemeliharaan jatuh pada "mengingat mendaftarkan setiap perintah
  berbahaya", padahal himpunan perintah berbahaya jauh lebih besar dan lebih
  sering bertambah daripada himpunan perintah aman.

Ini bertentangan dengan prinsip fail-safe yang sudah dianut di tempat lain:
`internal/driver/genericcli.Catalog` dengan zero value (`Curated: false`)
**selalu** mengembalikan `ClassDestructive` — vendor yang belum dikurasi
dianggap berbahaya secara default (lihat
`docs/adr/0004-generic-cli-driver-scrapligo.md`).

## Keputusan

**Klasifikasi perintah default ke `ClassDestructive`. Hanya perintah yang
cocok dengan pola BACA yang diketahui aman yang menjadi `ClassReadOnly`.**
Kita menyusun daftar putih (whitelist) perintah aman, bukan daftar hitam
perintah berbahaya.

Diterapkan sekarang pada `internal/driver/mikrotik` — tempat write op sedang
ditambahkan dan tersedia perangkat asli untuk diverifikasi. RouterOS punya
pola baca yang stabil dan mudah dikenali:

- Path yang berakhiran `/print` (semua perintah baca RouterOS: `/ppp/secret/print`,
  `/system/resource/print`, dst).
- Perintah observasi yang tidak mengubah state (`/ping`,
  `/interface/monitor-traffic` — himpunan `streamingBasePaths` yang sudah ada).

Selain itu → `ClassDestructive`. Dengan ini `destructivePaths` dihapus:
`/system/reboot`, `/ppp/secret/add`, `/ppp/profile/set`, `/ppp/active/remove`,
maupun path tak dikenal, semuanya destruktif tanpa perlu didaftarkan
satu per satu.

Konsekuensi khusus: operasi provisioning WRITE tidak akan pernah lewat jalur
auto-approve `ExecuteCommand`; hanya jalur ter-approve eksplisit
(`ExecuteCommandPreApproved`, dipakai adapter MCP setelah HITL) yang bisa
mengeksekusinya. `DecisionDeny` tetap memblokir keduanya.

## Konsekuensi

- **Positif:** lubang fail-open untuk write tertutup. Menambah operasi write
  baru aman secara default — tidak ada langkah "jangan lupa daftarkan sebagai
  destruktif" yang bisa terlewat. Beban pemeliharaan pindah ke himpunan yang
  lebih kecil dan lebih stabil (pola baca).
- **Negatif / batas:** perintah baca RouterOS yang TIDAK berakhiran `/print`
  dan bukan observasi terdaftar akan diklasifikasi destruktif (aman tapi
  menyusahkan — butuh approval untuk sesuatu yang read-only). Kalau kelak
  ada perintah baca semacam itu, tambahkan ke whitelist pola baca, bukan
  kembali ke daftar hitam.
- **Utang lintas-driver (belum diubah oleh ADR ini):** cisco, netconf, zteolt,
  huaweiolt, dan genieacs masih memakai posture fail-open lama. Tiap vendor
  punya pola baca sendiri (`show ...` di Cisco, `get`/`get-config` di NETCONF,
  task type di GenieACS), jadi migrasinya per-driver dan wajib diverifikasi ke
  perangkat/emulator masing-masing — bukan flip serentak tanpa uji. ADR ini
  menetapkan prinsipnya; migrasi tiap driver menyusul saat driver itu
  mendapat operasi write dan perangkat ujinya tersedia.
- **Kalau `command.Class` kelak diperkaya** (mis. kelas ketiga "write non-
  destruktif" dengan kebijakan approval berbeda), ADR ini di-supersede oleh
  ADR baru — bukan diedit di tempat.
