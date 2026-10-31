# Laporan Read-Only API e-Billing

- **Tanggal pengambilan:** 30 Oktober 2026
- **Base URL:** `https://api-service.ebilling.id/api`
- **Dokumentasi:** `https://api-service.ebilling.id/docs/api`
- **Tenant token:** `7315`
- **Metode:** hanya `GET`; tidak ada `POST`, `PUT`, `PATCH`, atau `DELETE`
- **Total request:** 64 request `GET`, seluruhnya `HTTP 200`
- **Token:** tidak ditulis di laporan ini; hanya disimpan sementara di luar repository

> Laporan ini berisi agregat data operasional. Nama pelanggan, nomor telepon, email,
> alamat, kredensial PPPoE, nomor rekening, dan URL invoice sengaja tidak dimasukkan.
> Respons mentah disimpan di `/tmp/opencode/ebilling/read` dengan permission terbatas
> dan tidak di-commit ke repository.

Daftar isi:

1. [Ringkasan eksekutif](#1-ringkasan-eksekutif)
2. [Request yang dilakukan](#2-request-yang-dilakukan)
3. [Profil tenant](#3-profil-tenant)
4. [Data pelanggan](#4-data-pelanggan)
5. [Pembayaran](#5-pembayaran)
6. [Tunggakan](#6-tunggakan)
7. [Status pelanggan](#7-status-pelanggan)
8. [Endpoint `/expired` dan data PPPoE](#8-endpoint-expired-dan-data-pppoe)
9. [Reseller](#9-reseller)
10. [Master data](#10-master-data)
11. [Temuan data quality](#11-temuan-data-quality)
12. [Rekomendasi](#12-rekomendasi)
13. [Lampiran](#13-lampiran)

---

## 1. Ringkasan eksekutif

| Metrik | Nilai |
|---|---:|
| Pelanggan tenant | 934 |
| Pelanggan berstatus aktif (`status=1`) | 733 (78,5%) |
| Pelanggan status `0` / `2` | 111 / 90 |
| Pembayaran tercatat (Oktober 2026) | 714 transaksi |
| Total nominal pembayaran | Rp70.570.000 |
| Pelanggan unik yang membayar | 671 |
| Pelanggan menunggak | 421 |
| Total tunggakan | Rp120.950.000 |
| Tunggakan periode 2024 | Rp74.000.000 (61,2%) |
| Tunggakan bulan berjalan (Oktober 2026) | Rp3.500.000 |
| Reseller terdaftar | 1 |
| Pelanggan di bawah reseller | 264 |
| Instalasi pelanggan baru | 0 |

Temuan utama:

1. **Sebagian besar tunggakan berasal dari periode lama.** Rp74.000.000 dari
   Rp120.950.000 (61,2%) adalah periode November–Desember 2024. Sebanyak 382
   pelanggan yang tercatat sudah membayar sampai Oktober 2026 masih membawa
   tunggakan 2024.
2. **Status pelanggan tidak konsisten.** 201 pelanggan (21,5%) berstatus
   `UNKNOWN`, tetapi seluruh 934 pelanggan memiliki `status_langganan = On`.
3. **Endpoint `/expired` bukan daftar pelanggan kedaluwarsa.** Endpoint ini
   mengembalikan 933 dari 934 pelanggan, semuanya `On`, dengan `expired_at`
   kosong.
4. **`status_pembayaran` tidak terisi.** Seluruh 714 transaksi pembayaran
   memiliki `status_pembayaran = null`.
5. **Metrik reseller perlu diverifikasi.** Dashboard menyebut 202 pelanggan
   aktif dan 0 non-aktif, sementara data pelanggan reseller berisi 202 aktif,
   39 status `2`, dan 23 status `0`.

---

## 2. Request yang dilakukan

| Endpoint | Request | Status | Data diterima |
|---|---:|---|---:|
| `GET /health` | 1 | 200 | `ok: true` |
| `GET /me` | 1 | 200 | 1 profil |
| `GET /hello-world` | 1 | 200 | 1 objek uji token |
| `GET /dashboard-mitra` | 2 | 200 | 2 ringkasan bulanan |
| `GET /resellers` | 1 | 200 | 1 reseller |
| `GET /resellers/2892/customers` | 3 | 200 | 264 pelanggan |
| `GET /resellers/2892/payments` | 2 | 200 | 189 pembayaran |
| `GET /lokasi` | 1 | 200 | 2 lokasi |
| `GET /tipe-pembayaran` | 1 | 200 | 8 paket |
| `GET /rekening` | 1 | 200 | 2 rekening |
| `GET /sales` | 1 | 200 | 4 user sales |
| `GET /instalasi-pelanggan-baru` | 1 | 200 | 0 |
| `GET /pelanggan` | 10 | 200 | 934 pelanggan |
| `GET /pelanggan?id_pelanggan=...` | 1 | 200 | 1 pelanggan |
| `GET /pelanggan-belum-bayar` | 5 | 200 | 421 pelanggan |
| `GET /pelanggan-belum-bayar?sampai=2026-10` | 1 | 200 | 421 pelanggan |
| `GET /status-pelanggan` | 10 | 200 | 934 baris status |
| `GET /status-pelanggan?status_langganan=Off` | 1 | 200 | 0 |
| `GET /expired` | 10 | 200 | 934 baris (933 unik) |
| `GET /pembayaran-pelanggan` | 8 | 200 | 714 pembayaran |
| `GET /pembayaran-pelanggan?bulan=2026-10` | 1 | 200 | 714 pembayaran |
| `GET /pelanggan/pppoe/959` | 1 | 200 | 1 pelanggan |
| **Total** | **64** | **200** | |

Catatan:

- `GET /resellers/{account_reseller}/customers` dan
  `GET /pelanggan/pppoe/{id_pelanggan}` berhasil dipanggil **hanya dengan path
  parameter**, tanpa query parameter duplikat yang tertulis di spec.
- Tidak ada request `POST /login` atau `POST /logout`; autentikasi memakai
  token Bearer yang diberikan pengguna.
- Tidak ada respons `401`, `403`, `404`, `422`, atau `429` pada pengambilan data.

---

## 3. Profil tenant

| Field | Nilai |
|---|---|
| `account` | `7315` |
| Level pengguna | Administrator |
| Status akun | `1` |
| `status_langganan` | `On` |
| Lokasi utama | `2782` |

Tenant memiliki:

- 1 reseller terdaftar: `account_reseller = 2892`
- 2 lokasi: `2782` dan `3934`
- 8 tipe pembayaran/paket
- 2 rekening pembayaran
- 4 user sales aktif
- 0 instalasi pelanggan baru

---

## 4. Data pelanggan

### 4.1 Distribusi status

| Field | Nilai | Jumlah | Persentase |
|---|---|---:|---:|
| `status` | `1` | 733 | 78,5% |
| `status` | `2` | 90 | 9,6% |
| `status` | `0` | 111 | 11,9% |
| `status_langganan` | `On` | 934 | 100% |
| `status_langganan` | `Off` | 0 | 0% |
| `status_pelanggan` | `ACTIVE` | 733 | 78,5% |
| `status_pelanggan` | `UNKNOWN` | 201 | 21,5% |

Cross-tab:

| `status` | `status_langganan` | `status_pelanggan` | Jumlah |
|---|---|---|---:|
| `1` | `On` | `ACTIVE` | 733 |
| `2` | `On` | `UNKNOWN` | 90 |
| `0` | `On` | `UNKNOWN` | 111 |

### 4.2 Distribusi paket

| Paket | Jumlah | Harga tercatat | Subtotal |
|---|---:|---:|---:|
| Paket Prem1 | 666 | Rp100.000 | Rp66.600.000 |
| Paket Reg1 | 190 | Rp100.000–Rp200.000 | Rp19.100.000 |
| Paket Prem2 | 31 | Rp150.000 | Rp4.650.000 |
| Paket Free | 25 | Rp0 | Rp0 |
| Paket 50rb | 13 | Rp50.000 | Rp650.000 |
| Paket Basic | 6 | Rp100.000 | Rp600.000 |
| Paket Prem3 | 2 | Rp200.000 | Rp400.000 |
| 30Mbps | 1 | Rp235.000 | Rp235.000 |
| **Total** | **934** | | **Rp92.235.000** |

Catatan: `Paket Basic` tidak ada di respons `GET /tipe-pembayaran`, tetapi 6
pelanggan memakai nama tersebut dengan `id_tipe_pembayaran` yang sama dengan
`Paket Prem1` (`16785`). Ini indikasi paket pernah di-rename.

### 4.3 Distribusi lokasi

| `id_lokasi` | Jumlah | Persentase |
|---|---:|---:|
| `2782` | 900 | 96,4% |
| `3934` | 34 | 3,6% |

### 4.4 Konektivitas dan kelengkapan data

| Field | Jumlah | Persentase |
|---|---:|---:|
| `jns_koneksi = PPPOE` | 889 | 95,2% |
| `jns_koneksi = 0` | 34 | 3,6% |
| `jns_koneksi = Static` | 9 | 1,0% |
| `jns_koneksi` kosong | 2 | 0,2% |
| `id_router` kosong | 2 | 0,2% |
| `tlp` terisi | 934 | 100% |
| `username` terisi | 934 | 100% |
| `email` terisi | 35 | 3,7% |
| `ppn = 0` | 934 | 100% |
| `jns_tagihan` kosong | 923 | 98,8% |
| `jns_tagihan = Pasca Bayar` | 11 | 1,2% |

### 4.5 Registrasi per tahun

| Tahun | Jumlah |
|---|---:|
| 2021 | 1 |
| 2022 | 127 |
| 2023 | 193 |
| 2024 | 230 |
| 2025 | 234 |
| 2026 | 149 |
| **Total** | **934** |

Rentang tanggal registrasi: `2021-06-08` sampai `2026-10-28`.

---

## 5. Pembayaran

### 5.1 Ringkasan

| Metrik | Nilai |
|---|---:|
| Transaksi | 714 |
| Total `nominal_pembayaran` | Rp70.570.000 |
| Total `nominal_harus_dibayar` | Rp70.570.000 |
| Pelanggan unik yang membayar | 671 |
| Pelanggan yang membayar periode Oktober 2026 | 668 |
| Transaksi dengan bukti pembayaran | 710 (99,4%) |
| `status_pembayaran` terisi | 0 |

### 5.2 Metode pembayaran

| Metode | Transaksi | Nominal |
|---|---:|---:|
| Cash | 601 | Rp58.970.000 |
| BNI | 113 | Rp11.600.000 |
| **Total** | **714** | **Rp70.570.000** |

### 5.3 Periode pembayaran

| Periode | Transaksi | Nominal | Pelanggan unik |
|---|---:|---:|---:|
| 2026-07 | 1 | Rp100.000 | 1 |
| 2026-08 | 8 | Rp800.000 | 8 |
| 2026-09 | 25 | Rp2.650.000 | 25 |
| 2026-10 | 669 | Rp65.970.000 | 668 |
| 2026-11 | 8 | Rp750.000 | 8 |
| 2026-12 | 2 | Rp200.000 | 2 |
| 2027-01 | 1 | Rp100.000 | 1 |
| **Total** | **714** | **Rp70.570.000** | **671** |

Seluruh transaksi memiliki `waktu_entry` pada Oktober 2026
(`2026-10-06` sampai `2026-10-30`), meskipun periode pembayarannya bervariasi.
Ini menunjukkan filter `bulan` pada endpoint pembayaran kemungkinan memakai
tanggal pencatatan, bukan `periode_pembayaran`.

### 5.4 Frekuensi pembayaran per pelanggan

| Jumlah pembayaran | Pelanggan |
|---:|---:|
| 1 kali | 642 |
| 2 kali | 21 |
| 3 kali | 4 |
| 4 kali | 3 |
| 6 kali | 1 |

Pelanggan tanpa record pembayaran pada dataset Oktober 2026:
**263 pelanggan (28,2%)**.

---

## 6. Tunggakan

### 6.1 Ringkasan

| Metrik | Nilai |
|---|---:|
| Pelanggan menunggak | 421 |
| Total `total_nominal_belum_lunas` | Rp120.950.000 |
| Jumlah `nominal_tagihan` (satu bulan per pelanggan) | Rp42.500.000 |
| Pelanggan dengan 1 periode tunggakan | 33 |
| Pelanggan dengan ≥ 2 periode tunggakan | 388 |

### 6.2 Tunggakan per tahun periode

| Tahun | Periode | Nominal | Porsi |
|---|---:|---:|---:|
| 2024 | 733 | Rp74.000.000 | 61,2% |
| 2025 | 211 | Rp21.150.000 | 17,5% |
| 2026 | 257 | Rp25.800.000 | 21,3% |
| **Total** | **1.201** | **Rp120.950.000** | **100%** |

Rincian 2024:

| Periode | Pelanggan | Nominal |
|---|---:|---:|
| 2024-11 | 363 | Rp36.650.000 |
| 2024-12 | 370 | Rp37.350.000 |

Tunggakan bulan berjalan:

| Periode | Pelanggan | Nominal |
|---|---:|---:|
| 2026-09 | 28 | Rp2.800.000 |
| 2026-10 | 35 | Rp3.500.000 |

### 6.3 Cohort berdasarkan periode tunggakan pertama

| Periode pertama | Pelanggan | Total tunggakan |
|---|---:|---:|
| 2024-11 | 363 | Rp106.150.000 |
| 2024-12 | 18 | Rp4.400.000 |
| 2025 | 13 | Rp6.100.000 |
| 2026 | 27 | Rp4.300.000 |
| **Total** | **421** | **Rp120.950.000** |

Cohort yang mulai menunggak November 2024 membawa **87,8%** dari seluruh
nilai tunggakan, meskipun sebagian dari nilai tersebut mencakup periode 2025
dan 2026.

### 6.4 Pembayaran terakhir vs tunggakan

| `periode_terakhir_dibayar` | Pelanggan | Nominal tunggakan |
|---|---:|---:|
| Oktober 2026 | 382 | Rp99.350.000 |
| September 2026 | 25 | Rp11.400.000 |
| Agustus 2026 | 7 | Rp4.400.000 |
| November 2026 | 2 | Rp500.000 |
| Juli 2026 | 2 | Rp1.400.000 |
| Februari 2026 | 1 | Rp1.700.000 |
| Maret 2025 | 1 | Rp2.000.000 |
| Desember 2026 | 1 | Rp200.000 |

**Anomali utama:** 382 pelanggan tercatat sudah membayar sampai Oktober 2026,
tetapi masih memiliki tunggakan lama. Nilai tunggakan mereka mencapai
Rp99.350.000 (82,1% dari total). Ini mengindikasikan data tagihan historis
yang belum direkonsiliasi, bukan tunggakan berjalan.

### 6.5 Jatuh tempo

| Tanggal jatuh tempo | Pelanggan |
|---:|---:|
| 01 | 17 |
| 02 | 65 |
| 03 | 64 |
| 05 | 158 |
| 10 | 115 |
| 15 | 1 |
| Kosong | 1 |

---

## 7. Status pelanggan

Endpoint `GET /status-pelanggan` mengembalikan 934 baris:

| `status_pelanggan` | Jumlah |
|---|---:|
| `ACTIVE` | 733 |
| `UNKNOWN` | 201 |

Seluruh 934 baris memiliki `status_langganan = On`. Filter
`GET /status-pelanggan?status_langganan=Off` mengembalikan **0 baris**.

Implikasi:

- `status_langganan` belum bisa dipakai untuk membedakan pelanggan aktif dan
  non-aktif pada tenant ini.
- Nilai `status = 0` dan `status = 2` dipetakan menjadi `UNKNOWN`, bukan
  `SUSPENDED` atau `DISMANTLE`.
- Dashboard reseller hanya menghitung `status = 1` sebagai pelanggan aktif,
  sehingga 39 pelanggan status `2` dan 23 pelanggan status `0` tidak terlihat
  sebagai non-aktif.

---

## 8. Endpoint `/expired` dan data PPPoE

| Metrik | Nilai |
|---|---:|
| Baris diterima | 934 |
| Pelanggan unik | 933 |
| Status `On` | 933 (100%) |
| `expired_at` terisi | 0 |
| `expired_at` kosong / `0` | 933 |
| `pppoe_user` terisi | 927 |
| `pppoe_password` terisi | 904 |
| `invoice` terisi | 933 |

Temuan:

- Endpoint ini mengembalikan hampir seluruh pelanggan (933 dari 934), bukan
  hanya pelanggan kedaluwarsa.
- `expired_at` kosong untuk semua baris, sehingga tidak bisa dipakai untuk
  menghitung masa berlaku.
- Data ini lebih tepat diperlakukan sebagai **inventaris kredensial PPPoE**
  daripada laporan expiry.
- 6 pelanggan tidak memiliki `pppoe_user` dan 29 pelanggan tidak memiliki
  `pppoe_password`.
- Respons memuat password PPPoE dan URL invoice; akses ke endpoint ini harus
  dibatasi dan tidak boleh dicatat ke log aplikasi.

---

## 9. Reseller

### 9.1 Profil reseller

| Field | Nilai |
|---|---|
| `account_reseller` | `2892` |
| Jumlah pelanggan | 264 |
| Lokasi | `3464` |
| Status `1` (aktif) | 202 |
| Status `2` | 39 |
| Status `0` | 23 |
| `status_langganan = On` | 264 |

### 9.2 Paket pelanggan reseller

| Paket | Jumlah |
|---|---:|
| Paket Basic | 237 |
| Paket Bronze | 13 |
| Paket Free | 6 |
| Paket Non Standar | 4 |
| paket 200 | 3 |
| PAKET 50 MBPS | 1 |

Koneksi: 260 PPPOE, 4 Hotspot. Rentang registrasi: `2022-10-01` sampai
`2026-10-01`.

### 9.3 Pembayaran reseller

| Metrik | Nilai |
|---|---:|
| Transaksi | 189 |
| Pelanggan unik yang membayar | 177 |
| Total nominal | Rp19.525.000 |
| Metode Cash | 163 |
| Metode BRI | 26 |

Periode pembayaran: 159 transaksi untuk Oktober 2026, 23 untuk September 2026,
4 untuk Agustus 2026, 2 untuk Juli 2026, dan 1 untuk Juni 2026.

### 9.4 Metrik dashboard

| Bulan | Revenue | Pelanggan | Reseller | Aktif | Non-aktif |
|---|---:|---:|---:|---:|---:|
| 2026-10 | Rp19.525.000 | 202 | 1 | 202 | 0 |
| 2026-09 | Rp19.325.000 | 202 | 1 | 202 | 0 |

Metrik reseller pada `GET /resellers`:

| Field | Nilai |
|---|---:|
| `total_pelanggan` | 202 |
| `total_lunas` | 189 |
| `total_belum_lunas` | 39 |
| `total_pemasukan` | Rp19.525.000 |
| `total_revenue` | Rp19.525.000 |
| `total_estimasi_pemasukan` | Rp20.575.000 |

Catatan konsistensi:

- `total_pelanggan = 202` sama dengan jumlah pelanggan `status = 1`.
- `total_belum_lunas = 39` sama dengan jumlah pelanggan `status = 2`.
- `total_lunas = 189` sama dengan jumlah **transaksi pembayaran**, bukan
  jumlah pelanggan unik yang membayar (177).
- Dashboard menyatakan `non_active_customer = 0`, padahal ada 39 pelanggan
  status `2` dan 23 pelanggan status `0` di data pelanggan reseller.
- Definisi `total_lunas`, `total_belum_lunas`, dan `non_active_customer` perlu
  dikonfirmasi ke penyedia API.

---

## 10. Master data

### 10.1 Lokasi

| `id_lokasi` | Tipe | Jumlah pelanggan tenant |
|---|---|---:|
| `2782` | POP | 900 |
| `3934` | POP | 34 |

### 10.2 Tipe pembayaran

| `id_tipe_pembayaran` | Nama | Harga | Jenis |
|---|---|---:|---|
| `53517` | 30Mbps | Rp235.000 | Wajib Tiap Bulan |
| `26629` | Tidak Ada | Rp0 | Tidak Wajib Tiap Bulan |
| `16789` | Paket Reg1 | Rp100.000 | Wajib Tiap Bulan |
| `16788` | Paket Free | Rp0 | Wajib Tiap Bulan |
| `16787` | Paket Prem3 | Rp200.000 | Wajib Tiap Bulan |
| `16786` | Paket Prem2 | Rp150.000 | Wajib Tiap Bulan |
| `16785` | Paket Prem1 | Rp100.000 | Wajib Tiap Bulan |
| `16784` | Paket 50rb | Rp50.000 | Wajib Tiap Bulan |

### 10.3 Rekening

Terdapat 2 rekening: DANA dan BNI. Nomor rekening dan nama pemilik tidak
ditampilkan.

### 10.4 Sales

Terdapat 4 user sales aktif: 2 berlevel Administrator dan 2 berlevel Teknisi.
Nama tidak ditampilkan.

### 10.5 Instalasi pelanggan baru

`GET /instalasi-pelanggan-baru` mengembalikan **0 data** pada saat pengambilan.

---

## 11. Temuan data quality

1. **Tunggakan legacy 2024 mendominasi.** Rp74.000.000 (61,2%) dari total
   tunggakan adalah periode November–Desember 2024. Sebanyak 382 pelanggan
   yang sudah membayar sampai Oktober 2026 masih membawa tunggakan ini.
   Rekonsiliasi atau write-off perlu dilakukan sebelum data dipakai untuk
   penagihan otomatis.
2. **Tunggakan berjalan jauh lebih kecil.** Tunggakan September–Oktober 2026
   hanya Rp6.300.000, sedangkan total tunggakan Rp120.950.000.
3. **Status pelanggan tidak konsisten.** 201 pelanggan berstatus `UNKNOWN`
   dengan `status_langganan = On`; filter `Off` selalu kosong. Perlu mapping
   status yang jelas antara `status`, `status_langganan`, dan
   `status_pelanggan`.
4. **`/expired` tidak bisa dipakai untuk expiry monitoring.** `expired_at`
   kosong untuk semua baris; endpoint justru mengembalikan hampir semua
   pelanggan.
5. **`status_pembayaran` tidak terisi.** Jangan mengandalkan field ini untuk
   menentukan status lunas; gunakan keberadaan transaksi atau nominal.
6. **Metrik reseller ambigu.** `total_lunas` dan `total_belum_lunas` tidak
   konsisten dengan jumlah pelanggan unik dan status pelanggan.
7. **Filter bulan pembayaran perlu diuji lebih lanjut.** Semua transaksi
   tercatat pada Oktober 2026, tetapi periode pembayaran mencakup Juli 2026
   sampai Januari 2027. Pastikan filter `bulan` memakai `waktu_entry`, bukan
   `periode_pembayaran`.
8. **Nama paket tidak konsisten.** `Paket Basic` muncul di data pelanggan,
   tetapi tidak ada di daftar tipe pembayaran; ID-nya sama dengan
   `Paket Prem1`.
9. **Kelengkapan kredensial PPPoE belum 100%.** 6 pelanggan tanpa
   `pppoe_user`, 29 tanpa `pppoe_password`.
10. **Data PII terekspos di beberapa endpoint.** `/pelanggan` memuat
    telepon/email/alamat, `/expired` memuat password PPPoE dan URL invoice,
    `/rekening` memuat nomor rekening. Endpoint ini perlu pembatasan akses
    dan redaksi log.

---

## 12. Rekomendasi

1. Lakukan rekonsiliasi tagihan 2024 sebelum memakai endpoint tunggakan untuk
   penagihan. Pisahkan "legacy arrears" dari "current arrears".
2. Minta vendor menjelaskan mapping `status` (`0`, `1`, `2`) dan definisi
   `status_langganan`, `total_lunas`, `total_belum_lunas`, serta
   `non_active_customer`.
3. Gunakan `/pelanggan/pppoe/{id}` untuk mengambil kredensial satu pelanggan
   secara terkontrol; jangan menyimpan hasil `/expired` secara massal.
4. Untuk integrasi Polyglot, buat adapter read-only terlebih dahulu:
   - sync pelanggan dari `/pelanggan`
   - sync pembayaran dari `/pembayaran-pelanggan`
   - gunakan `/pelanggan/pppoe/{id}` untuk data PPPoE
   - jangan memakai `/expired` sebagai sumber status expiry
5. Tambahkan guard di adapter agar hanya mengizinkan method `GET` sampai
   proses rekonsiliasi selesai.
6. Jangan menulis token, password PPPoE, nomor rekening, atau PII ke log,
   database aplikasi, atau file yang di-commit.

---

## 13. Lampiran

- Spesifikasi OpenAPI hasil scraping: `/tmp/opencode/ebilling/openapi.json`
- Ringkasan endpoint: `/tmp/opencode/ebilling/endpoints.md`
- Respons mentah (read-only): `/tmp/opencode/ebilling/read/`
- Token tidak disimpan di repository; file token sementara sudah dihapus
  setelah pengambilan data selesai

Endpoint yang **tidak** dipanggil karena bersifat write/update:

- `POST /login`
- `POST /logout`
- `POST /pelanggan`
- `POST /pembayaran-pelanggan`
- `POST /pembayaran-pelanggan-2809`

Tidak ada perubahan data yang dilakukan pada API e-Billing.
