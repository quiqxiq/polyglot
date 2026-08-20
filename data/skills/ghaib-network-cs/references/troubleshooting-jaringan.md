# Troubleshooting Jaringan — Panduan Lengkap

Dipanggil dari `SKILL.md` untuk semua keluhan teknis: internet mati total, internet lambat, dan masalah WiFi.

## Prinsip Umum

- **Satu langkah per pesan** — kirim satu instruksi, tunggu balasan/hasil dari pelanggan, baru lanjut ke langkah berikutnya. Jangan tumpuk banyak instruksi sekaligus.
- Kalau indikasi gangguan sudah jelas di jalur fisik (lampu LOS merah), **jangan** minta pelanggan restart modem berulang kali — itu tidak akan memperbaiki masalah jalur fiber.
- Jangan mengulang langkah yang sama lebih dari sekali dalam satu sesi.
- Kalau semua langkah relevan sudah dicoba dan belum solve → langsung tawarkan jadwal kunjungan teknisi atau eskalasi. Lihat `eskalasi-dan-faq.md`.

---

## A. Tidak Ada Koneksi Internet Sama Sekali

### Langkah 1 — Cek Lampu Indikator ONT/Modem

Minta pelanggan sebutkan kondisi lampu di alat ONT/modem mereka:

- **PWR (Power)** — harus menyala tetap.
- **PON / LOS (Loss of Signal)** — harus menyala hijau tetap, bukan berkedip/menyala merah.
- **LAN** — menyala kalau kabel LAN ke router/PC terpasang dan aktif.
- **WLAN/WiFi** (kalau ONT sekaligus jadi router WiFi) — menyala kalau WiFi aktif.

Interpretasi hasil:

- **PWR mati total** → minta pelanggan cek colokan listrik, coba pindah ke stop kontak lain, pastikan adaptor tidak longgar/rusak. Kalau tetap mati setelah itu → kemungkinan alat rusak, jadwalkan kunjungan teknisi.
- **LOS berkedip atau menyala merah** → ini indikasi gangguan di jalur fiber optik, **bukan** masalah setting/perangkat pelanggan. Restart modem tidak akan memperbaikinya. Sampaikan ke pelanggan bahwa ini kemungkinan gangguan jalur, dan langsung arahkan ke jadwal pengecekan teknisi (lihat `eskalasi-dan-faq.md`) — tidak perlu lanjut ke Langkah 2–4.
- **Semua lampu normal tapi tetap tidak ada internet** → lanjut ke Langkah 2.

### Langkah 2 — Restart Modem/ONT

1. Cabut adaptor listrik ONT/modem (bukan cuma tombol power, cabut dari sumber listriknya).
2. Tunggu 30 detik.
3. Colok kembali, tunggu 2–3 menit sampai semua lampu kembali stabil normal.
4. Tanyakan apakah internet sudah kembali.

### Langkah 3 — Cek Cakupan Masalah

Tanyakan: masalah terjadi di **semua perangkat** (HP, laptop, TV — semua kena) atau **cuma satu perangkat saja**?

- **Cuma satu perangkat** → kemungkinan masalah ada di perangkat itu sendiri (driver WiFi, kabel LAN rusak, pengaturan jaringan perangkat), bukan gangguan dari sisi Ghaib Network. Arahkan pelanggan cek/restart perangkat tersebut, lupakan jaringan WiFi lalu sambung ulang, atau cek kabel LAN-nya — bukan eskalasi ke jaringan.
- **Semua perangkat kena** → lanjut ke Langkah 4.

### Langkah 4 — Cek Router Terpisah (kalau ONT dan router beda alat)

Kalau pelanggan pakai router WiFi terpisah dari ONT:

1. Cek juga lampu indikator router (power, WAN/internet).
2. Restart router juga (cabut adaptor, tunggu 30 detik, colok lagi, tunggu 2 menit).
3. Pastikan kabel LAN dari ONT tersambung ke port **WAN** router (bukan salah satu port LAN biasa).

### Langkah 5 — PPPoE / Gagal Login (kalau pelanggan pakai username-password internet)

Kalau router menampilkan status "connecting"/"authenticating" terus-menerus, atau ada notifikasi gagal login PPPoE:

1. Pastikan username & password PPPoE yang diinput sesuai persis dengan yang diberikan saat pemasangan (huruf besar-kecil berpengaruh).
2. Kalau pelanggan yakin sudah benar tapi tetap gagal → ini kemungkinan bukan salah setting, melainkan status akun (misalnya akun sedang diisolir karena tagihan menunggak). Cek dulu status tagihan pelanggan lewat `tagihan-dan-pembayaran.md` — kalau memang menunggak, ini penyebabnya, arahkan ke alur pembayaran.
3. Kalau tagihan aman tapi tetap gagal login → eskalasi ke tim teknis, sertakan info bahwa ini masalah otentikasi bukan gangguan jalur.

### Langkah 6 — Belum Solve

Kalau sudah melewati semua langkah relevan di atas dan belum berhasil → jangan mengulang langkah yang sama, langsung tawarkan jadwal kunjungan teknisi atau eskalasi (lihat `eskalasi-dan-faq.md`). Jangan janjikan waktu perbaikan pasti kalau belum ada kepastian dari tim teknis.

---

## B. Internet Lambat

### Langkah 1 — Cakupan Masalah

Tanyakan: lambat di **semua perangkat** atau **cuma sebagian**? Lambat di **semua aplikasi/situs** atau cuma yang tertentu?

- Kalau cuma situs/aplikasi tertentu yang lambat sementara yang lain normal → kemungkinan masalah ada di server/aplikasi tersebut, bukan di jaringan Ghaib Network.

### Langkah 2 — Jam Pemakaian

Kalau kejadian di jam sibuk (umumnya sekitar pukul 19:00–23:00) dan paket pelanggan sifatnya shared bandwidth, kondisi sedikit lebih lambat di jam tersebut masih dalam batas wajar. `[ISI: sebutkan kalau paket Ghaib Network sifatnya dedicated (bukan shared), sesuaikan penjelasan di sini]`

### Langkah 3 — Speed Test

Minta pelanggan melakukan speed test (mis. lewat speedtest.net atau aplikasi sejenis) dengan kondisi:

- Idealnya pakai kabel LAN langsung ke laptop/PC (bukan WiFi) untuk hasil paling akurat.
- Tidak ada perangkat lain di jaringan yang sama yang sedang download/streaming berat saat pengujian.

Minta pelanggan kirimkan hasilnya (angka download/upload dalam Mbps).

### Langkah 4 — Bandingkan dengan Paket

Bandingkan hasil speed test dengan kecepatan paket yang dilanggan pelanggan (cek nama paketnya di `paket-dan-harga.md` kalau perlu).

- **Hasil jauh di bawah paket** (misal kurang dari separuh kecepatan langganan) dan bukan di jam sibuk → catat hasil speed test beserta jam pengujian, eskalasi ke tim teknis level 2.
- **Hasil mendekati kecepatan paket** tapi pelanggan tetap merasa lambat → kemungkinan bottleneck di WiFi/perangkat lokal, lanjut ke bagian C (Masalah WiFi) di bawah.

---

## C. Masalah WiFi (kabel/ONT normal, tapi WiFi bermasalah)

1. Cek jarak dan penghalang — tembok beton tebal, banyak lantai/sekat, atau jarak yang terlalu jauh dari router bisa melemahkan sinyal.
2. Restart router/ONT (cabut adaptor 30 detik, colok lagi).
3. Tanyakan jumlah perangkat yang terkoneksi bersamaan — terlalu banyak perangkat aktif sekaligus bisa membuat jaringan terasa berat.
4. Kalau perangkat pelanggan mendukung, cek apakah WiFi memakai frekuensi 2.4GHz atau 5GHz — 5GHz lebih cepat tapi jangkauannya lebih pendek dan mudah terhalang tembok; 2.4GHz jangkauannya lebih jauh tapi lebih lambat dan lebih rawan interferensi dari perangkat lain. Sarankan pindah frekuensi sesuai kebutuhan pelanggan.
5. Kalau tetap bermasalah, tawarkan **reset factory** sebagai langkah terakhir — beri tahu pelanggan ini akan menghapus nama & password WiFi custom yang sudah diatur sebelumnya, dan mereka perlu setting ulang setelahnya.
6. Kalau rumah/area pelanggan luas dan sinyal WiFi memang tidak menjangkau semua ruangan meski alat berfungsi normal → ini bukan gangguan teknis, informasikan opsi tambahan seperti access point/extender/mesh WiFi. `[ISI: apakah Ghaib Network menyediakan layanan tambah access point berbayar — sebutkan detail & harganya kalau ada, atau hapus baris ini kalau tidak tersedia]`

---

## D. Kapan Berhenti Troubleshooting Sendiri & Eskalasi

Jangan mengulang langkah yang sama lebih dari sekali. Setelah semua langkah relevan di atas dicoba dan belum solve:

- Tawarkan jadwal kunjungan teknisi, **atau**
- Eskalasi ke tim teknis/NOC dengan ringkasan: jenis perangkat yang dipakai pelanggan, langkah-langkah yang sudah dicoba, hasil speed test (kalau ada), dan status lampu indikator.

Lihat kriteria & cara eskalasi lengkap di `references/eskalasi-dan-faq.md`.
