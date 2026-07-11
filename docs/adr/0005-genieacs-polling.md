# 0005 — GenieACS Driver Memakai Polling (Bukan Push)

## Status
Diterima

## Konteks

Prinsip umum proyek ini adalah "tidak ada polling" — driver device idealnya
memakai koneksi persisten yang mem-push event (mis. Mikrotik dengan dua
koneksi exec/stream, lihat `docs/adr/0003-mikrotik-dual-connection-streaming.md`).
Polling memboroskan request, menambah latensi rata-rata, dan menyembunyikan
kegagalan di balik interval.

GenieACS adalah pengecualian yang dipaksa oleh desain sistem eksternalnya,
bukan oleh pilihan kita:

1. **Model eksekusi NBI GenieACS asinkron via antrian.** `POST /devices/{id}/tasks`
   mengembalikan `202 Accepted` dengan task yang MASUK ANTRIAN — bukan hasil
   eksekusi. Task baru benar-benar dijalankan saat CPE melakukan sesi TR-069
   berikutnya (inform), yang waktunya di luar kendali NBI.

2. **NBI tidak punya channel push/WebSocket untuk penyelesaian task.**
   Satu-satunya cara mengetahui sebuah task sudah dieksekusi adalah
   me-query ulang `GET /tasks/?query={"_id":"..."}` dan melihat apakah task
   sudah hilang dari antrian (array kosong = sudah dieksekusi). Tidak ada
   endpoint "tunggu sampai selesai", tidak ada callback, tidak ada
   long-poll resmi.

Kontrak `port.DeviceDriver.Execute` bersifat blocking (mengembalikan hasil
saat command selesai). Untuk memenuhi kontrak itu di atas model antrian yang
hanya bisa di-query, driver HARUS menjembatani "asinkron + query-only" menjadi
"blocking" — dan satu-satunya jembatan yang tersedia adalah polling.

## Keputusan

`internal/driver/genieacs`'s `waitForTask` melakukan polling
`GET /tasks/?query={"_id":"<taskID>"}` pada interval `d.pollInterval`
(default `DefaultPollInterval`, dapat di-override lewat
`Target.Extra["poll_interval"]`) sampai task hilang dari antrian atau `ctx`
berakhir. Hanya `ctx` yang mengakhiri loop; error request transien
ditelan dan di-retry, konsisten dengan kontrak blocking-Execute.

Ini ditandai `// DEVIASI: polling` di `waitForTask` karena menyimpang dari
prinsip "no polling" — deviasi yang sah dan didokumentasikan di sini, bukan
kelalaian.

## Konsekuensi

- **Positif:** driver memenuhi kontrak `port.DeviceDriver` yang blocking
  tanpa mengubah kontrak itu demi satu vendor. Interval yang dapat
  dikonfigurasi membuat operator bisa menukar latensi vs beban NBI per
  device.
- **Negatif:** ada beban request berkala selama sebuah task pending, dan
  latensi penyelesaian yang terlihat dibatasi bawah oleh `pollInterval`.
  Keduanya intrinsik pada model NBI GenieACS, bukan artefak implementasi
  kita.
- **Kalau GenieACS kelak menyediakan push** (mis. via ekstensi atau
  komponen di depan NBI), ADR ini di-supersede oleh ADR baru dan
  `waitForTask` diganti — bukan diedit di tempat.
