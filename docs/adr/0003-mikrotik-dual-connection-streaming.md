# 0003 — Mikrotik: Dua Koneksi Persisten (Exec + Stream), Bukan Satu

## Status
Diterima

## Konteks

Mikrotik RouterOS API mendukung command yang balasannya datang berkali-kali
sampai dibatalkan secara eksplisit (`/ping`, `/interface/monitor-traffic`,
dan `/.../print` dengan `follow`/`follow-only`/`interval=`) — bukan cuma
command sekali-jalan yang berakhir dengan `!done`.

Dicoba dengan SATU `*routeros.Client` untuk keduanya, gejala yang muncul:
begitu ada command streaming aktif, command execute lain di connection yang
sama jadi tidak bisa jalan (macet/tidak dapat balasan). Ini dikonfirmasi
langsung dari source `github.com/go-routeros/routeros/v3` (bukan tebakan):

1. **Mode sync (default, tanpa `Async()`)**: `RunArgsContext` dalam mode ini
   (`runArgsContextSync`) langsung memanggil `ReadSentence()` di reader yang
   sama, TANPA pengecekan tag sama sekali dan TANPA memeriksa `ctx` — kalau
   ada `Listen` lain yang juga sedang membaca dari reader yang sama secara
   bersamaan, keduanya berebut baca sentence berikutnya dari wire yang sama.
   Tidak ada isolasi antar command sama sekali di mode ini.

2. **Mode async**: `ListenArgsQueueContext` OTOMATIS mengaktifkan `Async()`
   kalau belum aktif — dan setiap kali dipanggil (baik untuk `Run` maupun
   `Listen`), context yang dilewatkan didaftarkan ke sebuah goroutine:
   `go func() { <-ctx.Done(); c.r.Cancel() }()`. `c.r.Cancel()` ini
   membatalkan **reader untuk SELURUH koneksi**, bukan cuma command/stream
   itu saja. Konsekuensinya: kalau context yang dioper ke satu panggilan
   `RunArgsContext`/`ListenArgsContext` itu timeout atau di-cancel oleh
   pemanggilnya, SELURUH koneksi (termasuk command/stream lain yang sedang
   berjalan di situ) ikut mati.

3. Proses penulisan sentence (`BeginSentence`/`WriteWord`/`EndSentence`) di
   kedua jalur tidak terlihat dilindungi mutex sepanjang durasinya — tidak
   ada jaminan eksplisit dari source bahwa dua pemanggilan `Run`/`Listen`
   yang bersamaan pada satu `*Client` aman dari byte yang saling menyisip
   di wire.

## Keputusan

1. **Dua `*routeros.Client` independen per device**: `exec` (khusus
   `Execute`, lewat `RunArgsContext`) dan `stream` (khusus `Stream`, lewat
   `ListenArgsContext`). Tidak pernah berbagi satu koneksi. Ini menghilangkan
   masalah #1 dan #3 di atas secara struktural — dua TCP connection berbeda
   tidak mungkin berebut reader/writer yang sama.

2. **Context yang dioper ke pustaka go-routeros SELALU `context.Background()`,
   TIDAK PERNAH context milik caller (`ctx` parameter `Execute`/`Stream`).**
   Ini menghindari masalah #2: koneksi persisten tidak boleh mati hanya
   karena satu request individual timeout. `ctx` milik caller dipakai murni
   untuk mengontrol berapa lama `Execute`/`Stream` (fungsi Go biasa, bukan
   panggilan ke pustaka) mau MENUNGGU hasil — lewat goroutine terpisah +
   `select` melawan `ctx.Done()`, bukan dioper langsung ke pustaka.
   Konsekuensi: kalau caller menyerah menunggu, operasi yang sebenarnya
   (jalan di `context.Background()`) tetap dibiarkan selesai sendiri di
   background — tag-nya "ditinggal", bukan dipaksa mati, supaya reader
   koneksi tidak ikut rusak.

3. **Serialisasi wire per koneksi lewat semaphore (`chan struct{}` kapasitas
   1), bukan `sync.Mutex` biasa** — supaya menunggu giliran pun tetap bisa
   dibatalkan lewat `ctx` caller (`select` antara kirim ke semaphore vs
   `ctx.Done()`), sesuatu yang tidak bisa dilakukan `sync.Mutex.Lock()`
   secara native. Untuk `Execute`, semaphore ditahan sepanjang durasi
   `RunArgsContext` yang sebenarnya (dilepas oleh goroutine background,
   BUKAN oleh pemanggil yang mungkin sudah menyerah duluan) — supaya
   panggilan `Execute` berikutnya tidak ikut menulis ke wire sebelum
   panggilan sebelumnya benar-benar selesai. Untuk `Stream`, semaphore
   hanya ditahan sebentar saat MENERBITKAN `ListenArgsContext` (yang
   menurut source pustakanya langsung kembali, tidak menunggu data) —
   bukan sepanjang umur stream, supaya beberapa stream (mis. `/ping` dan
   `/interface/monitor-traffic` bersamaan) tetap bisa berjalan paralel di
   satu koneksi stream yang sama.

4. **`Client.Queue` diisi angka besar (1000), bukan dibiarkan default (0)**.
   Default 0 berarti channel `Listen()` unbuffered — dispatcher `asyncLoop`
   (satu goroutine yang melayani SEMUA tag di satu koneksi) akan macet
   menunggu konsumen kita membaca setiap baris sebelum lanjut memproses
   sentence berikutnya, yang berarti satu stream yang lambat dikonsumsi
   bisa menahan tag lain di koneksi yang sama. Buffer besar menghindari ini.

5. **Auto-reconnect per koneksi** (lihat `connect.go`): masing-masing
   `exec`/`stream` diawasi goroutine tersendiri yang mendengarkan channel
   error dari `Async()`; begitu terputus, redial dengan backoff sampai
   berhasil atau `Driver.Close()` dipanggil. Tidak ada polling status
   koneksi — murni event-driven lewat channel error itu.

## Konsekuensi

- `port.DeviceDriver` **tidak** ditambahi method `Stream` — itu tetap
  ditunda sesuai ADR 0002. Sebagai gantinya, `port.StreamingDeviceDriver`
  (interface baru, terpisah) jadi kemampuan opt-in per vendor. Vendor lain
  (cisco, genericssh, netconf, zteolt, huaweiolt, genieacs) tidak perlu
  method stub tambahan apa pun.
- `command.Result` bertambah field `Rows []map[string]string` (menggantikan
  kebutuhan akan field flat tunggal) — supaya balasan multi-baris (mis.
  `/interface/print` tanpa filter) tidak diam-diam kehilangan baris selain
  yang pertama, dan supaya satu event stream (satu `!re`) punya bentuk yang
  sama dengan satu baris hasil `Execute`.
- Paket `internal/driver/mikrotik` punya lebih dari dua file
  (`driver.go`, `connect.go`, `stream.go`, `commands.go`, `errors.go`) —
  DEVIASI tercatat dari baseline dua-file di CLAUDE.md §1.2, karena
  kompleksitas dual-connection + reconnect + streaming yang nyata
  membutuhkannya. `commands.go` tetap hanya berisi katalog command +
  klasifikasi risiko + helper konversi, sesuai perannya yang asli.
- **Siapa pun yang mengubah kode ini di masa depan JANGAN mengoper `ctx`
  milik caller langsung ke `RunArgsContext`/`ListenArgsContext`** — itu
  akan mengembalikan persis bug yang ADR ini selesaikan (koneksi persisten
  mati karena satu request timeout). Kalau butuh perilaku itu benar-benar
  disengaja (misalnya arsitektur berubah jadi satu koneksi sekali pakai per
  command), itu keputusan baru yang butuh ADR baru, bukan revisi diam-diam
  ke ADR ini.
