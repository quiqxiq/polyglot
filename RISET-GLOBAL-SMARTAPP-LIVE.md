# Riset Implementasi Global SmartApp — Analisis Live Admin Panel (Real ISP)

**Tanggal riset:** 31 Oktober 2026  
**Target:** `https://app.global-smartapp.com` (panel admin/partner)  
**Metode:** akses terautentikasi memakai akun mitra yang diberikan pengguna, **read-only** (GET halaman + inventarisasi form/endpoint via HTML & inline JS), tanpa aksi tulis/delete/approve, rate-limited ±0,6–1 detik/permintaan.  
**Cakupan:** 69 halaman admin dipetakan, 273 endpoint HTTP diinventarisasi, seluruh menu (47 item, termasuk submenu) dianalisis.  
**Etika/keamanan:** kredensial tidak ditulis ke laporan; sesi diakhiri dengan logout; file cookie & secret TOTP dihapus dari mesin riset; tidak ada upaya akses lintas mitra, bypass otorisasi, atau perubahan data.

> **Batasan penting:** akun mitra yang diberikan **tidak berisi data operasional** (dashboard menampilkan 0 pelanggan; tabel kosong). Yang diamati adalah **permukaan implementasi** (struktur menu, form/field, alur UI, endpoint, mekanisme job, integrasi) — **bukan** perilaku/angka billing produksi milik ISP lain. Tidak ada akses source code, database, atau konfigurasi server.

Halaman login admin memunculkan **setup 2FA (TOTP)**. Untuk melanjutkan riset dipilih opsi bawaan aplikasi **“Ingatkan Nanti”** (tidak mengaktifkan maupun menonaktifkan 2FA; tidak mengubah setelan keamanan akun). Pada login berikutnya halaman setup akan muncul kembali.

---

## 1. Ringkasan Eksekutif (12 Temuan Kunci)

1. **Monolit PHP 7.4 bergaya CodeIgniter**, server-side rendered + jQuery/Bootstrap (template Skydash), AJAX/DataTables server-side; **bukan SPA/API-first**. Bukti: `ci_session`, `ci_csrf_token`, `x-powered-by: PHP/7.4.33`, Cloudflare di depan.
2. **Domain sangat luas dalam satu panel**: billing ISP, jaringan (MikroTik/OLT/ACS/GIS), SDM (absensi/karyawan), kolektor, agen hotspot, inventory, PSB, notifikasi, ticketing, laporan pajak.
3. **Model billing berbasis masa aktif (expired-driven)**, khas RT/RW Net: pembayaran menggeser tanggal expired layanan; “invoice” melekat pada pelanggan/langganan, bukan menu invoice terpisah. Ada mode **Prabayar / Pascabayar**, **Tanggal Tutup Buku 1–28**, dan setelan **“Hari Belum Bayar”** (default H-7, untuk tampilan/filter).
4. **Billing engine cukup lengkap**: PPN (global mitra + per pelanggan), diskon Rp/%, **prorata + denda keterlambatan**, **biaya tambahan langganan berulang**, dan **hutang-piutang manual** per pelanggan.
5. **Pembayaran**: konfirmasi manual dengan **review bukti transfer** (Order ID, tanggal, via transfer, bukti, biaya tambahan, diskon, PPN, nominal) + input income kasir; **mass payment** per batch memakai **job queue background**.
6. **Payment gateway**: pilihan **Global Moneta (aggregator, direkomendasikan)**, **Midtrans**, atau tanpa gateway; konfigurasi bank/VA, **fee admin 100% mitra / 100% pelanggan / 50-50**, tampil/tidak di invoice, kredensial sandbox & production, dan **PIN WD** (withdrawal).
7. **OLT onboarding berupa wizard multi-step** yang mengotomasi tunnel & firewall MikroTik (L2TP / WireGuard ROS7), port publik, NAT, dan penyimpanan; mendukung OLT **CARD** (telnet+SNMP+web) dan **BUNDLE**, brand ZTE/Huawei/HSGQ/HiOSO/VSOL/C-Data.
8. **Sinkronisasi OLT 3 jalur**: **Upload DAT**, **DAT Telnet** (via *Server Global* atau *MikroTik Local*), dan **Sync Cerdas** (step-by-step) — sesuai pola `ANALISIS_FITUR_GLOBAL_SMARTAPP.md` §3.2.C.
9. **ACS terpusat** di `http://acsmonitor.global-smartapp.com:321` dengan kredensial ACS global yang ditampilkan di UI; fitur *summon/refresh*, cari ONT by **SN/MAC**, tambah/bind ke pelanggan, **reboot/reset factory/delete** ONT.
10. **GIS/topologi nyata**: ODC, ODP (parent ODC/ODP, jumlah port, client terhubung per tipe layanan), joint closure, tiang, kalkulasi rasio (splitter/jalur/8 ODP, “sisa laser”), estimasi redaman, mapping HERE Maps + import **KMZ**.
11. **Kolektibilitas & kanal adalah warga kelas satu**: **Collector** + **Setoran Collector** (sisa di tangan), **Agent Hotspot** (voucher, komisi, hutang voucher, setoran, konfirmasi voucher), **absensi karyawan** (selfie + GPS + izin/sakit + rekap) — sebagian **dikunci paket langganan**.
12. **Notifikasi multi-kanal**: FCM push + WhatsApp gateway + Telegram bot; segmentasi audiens (semua/ODP/server/expired/tipe user/PPPoE offline/custom), template pesan, **jadwal auto-notif per jam**, histori terkirim/gagal. Modul **WhatsApp Gateway & Add Ons berstatus “Akses Terbatas”** (harus upgrade paket); **Absensi “Fitur Tidak Tersedia”** untuk akun ini → produk memakai **subscription/add-on gating**.

---

## 2. Stack & Arsitektur Teknis (observasi)

| Aspek | Temuan |
|---|---|
| Bahasa/framework | PHP 7.4.33 (`x-powered-by`), pola **CodeIgniter** (`ci_session`, `ci_csrf_token`, URL controller/method) |
| Frontend | Server-rendered HTML + jQuery 3.6, Bootstrap 4.6, template **Skydash**, SweetAlert2, Bootstrap-select, **DataTables 1.12 server-side**, Highcharts, HERE Maps |
| Async/job | **Job queue internal** lewat `router_job/step`, `status`, `ack`, `stop` (batch async: mass payment, import voucher, dsb.) + background HTTP `sending_wa_gateway_background[/stop]` |
| Notifikasi | Firebase Cloud Messaging (FCM) push; WhatsApp gateway device/session; Telegram Bot API |
| Integrasi eksternal | Payment gateway **Global Moneta** & **Midtrans**; ACS server terpisah (`:321`); HERE Maps; Google Analytics; Cloudflare |
| Sesi/keamanan | Cookie `ci_session` (HttpOnly, Secure, ±2 jam), 2FA TOTP pasca-login (opsional/postponable), CSRF token form, role & permission granular; tidak terlihat HSTS/CSP/X-Frame pada respons halaman admin |
| Multi-tenant | Satu akun = satu **mitra/owner**; ID mitra tampak terenkode di URL/tautan, data terpisah per mitra |
| Data interchange | Server-rendered tabel + AJAX JSON; export **Excel/PDF** (income, pengeluaran, absensi, laporan); import **DAT** (backup OLT) dan **KMZ** (peta) |

---

## 3. Peran, Permission & Gating

- **Role bawaan** (dropdown form karyawan): `Admin`, `Teknisi`, `Collector`, `Sales`; di level akun ada **Owner**; selain itu ada **Agent Hotspot** (akun agen voucher) dan **Collector** sebagai entitas operasional.
- **Permission granular per modul & aksi** (form Tambah Karyawan): `permission[Pelanggan][]`, `permission[Report][]`, `permission[Inventory][]`, `permission[Income][]`, `permission[Collector][]`, `permission[Agent_Hotspot][]`, `permission[Planning_PSB][]`, `permission[Action][OLT]`, `permission[Action][Topology]`, serta `dashboard_filter[]`.
- **Gating langganan**: halaman `Add Ons` & `WhatsApp Gateway` berjudul **“Akses Terbatas”** dengan pesan upgrade paket; halaman **Absensi** menampilkan **“Fitur Absensi Tidak Tersedia”** → modul dijual sebagai add-on/paket.
- **Menu sidebar** dirender berdasarkan izin; halaman tanpa izin balik 403 “Anda tidak memiliki izin untuk mengakses halaman ini”.

---

## 4. Information Architecture (menu utama)

- **Monitoring**: Dashboard (`/admin`) — kartu TOTAL / SUDAH BAYAR / BELUM BAYAR / ISOLIR / PSB, grafik income & pengeluaran, info MikroTik (CPU/RAM/HDD/board/ROS/uptime), peta.
- **Jaringan**: Mikrotik (List & Tambah), OLT (List & Tambah), Topology (Server List, ODC List, ODP List, Joint Closure, Mapping All Point, Topologi Jaringan), Calculator Ratio (Splitter / 8 ODP / Jalur Percabangan), IP Publik, Burst Limit Calc.
- **Billing & pelanggan**: Set Page Isolir, Banner, Discount, Prorata, GlobalACS, Remote Device, Pelanggan (`/user_list`), Planning PSB, Pembayaran, Income, Pengeluaran, Hutang Piutang, Biaya Tambahan Langganan, Payment Gateway, Payment Method manual.
- **Operasional lapangan**: Karyawan, Absensi (Saya/Rekap/Pengaturan), Collector (List/Setoran), Agent Hotspot (List/Confirm Voucher/Bagikan Voucher/Tagihan/Setoran).
- **Inventory**: Master Barang, Barang Datang, Barang Keluar, Stock Opname, Modem/ONT, Tutorial.
- **Akses & komunikasi**: Telegram Bot, WhatsApp Gateway, APK Client, Notification Apps.
- **Laporan**: Report Summary, Report Penjualan, Report PSB, Report Tax, Report Sales, Log Activity.
- **Akun**: My Profile (profil mitra, PPN, metode pembayaran, tutup buku, hari belum bayar, logo, ganti password, 2FA, sync income, tambah client/wilayah), Logout.

> Rincian 69 halaman ada di **Lampiran A**; inventaris 273 endpoint di **Lampiran B**.

---

## 5. Auth, Sesi & 2FA

- Login tunggal `POST /auth` (`email`, `password`, `ci_csrf_token`); sukses → `/admin`, tetapi bila 2FA belum disetel muncul `/two_factor_setup`.
- Setup 2FA: QR TOTP + secret base32 ditampilkan, verifikasi kode 6 digit via `POST /two_factor/activate`; tersedia `POST /two_factor/later` (“Ingatkan Nanti”) dan `POST /two_factor/decline` (menonaktifkan 2FA dengan pernyataan persetujuan).
- Sesi: cookie `ci_session` HttpOnly/Secure ±2 jam; hidden `ci_csrf_token` ada tetapi nilainya kosong pada halaman yang diamati (kemungkinan CSRF nonaktif/diisi JS) — perlu verifikasi lebih lanjut.
- Logout: `GET /auth/logout` → kembali ke `/auth`.
- Portal publik: `/auth/client` (login pelanggan), `/auth/registration` (registrasi mitra baru), `/forget_password` → `POST /fpass`; saat riset halaman publik menampilkan pesan **maintenance**.

---

## 6. Implementasi per Domain

### 6.1 Dashboard & Monitoring
- Endpoint kunci: `get_info_dashboard/{id_mikrotik}` (CPU/RAM/HDD/board/ROS/uptime), `get_traffic`, `get_traffic_income`, `get_traffic_pengeluaran`, `getlog`, `data_interface_hotspot`, `endpoint`, `check_ping`, `get_odp_points`, `disconnect_mikrotik_server`.
- Kartu ringkas: TOTAL / SUDAH BAYAR / BELUM BAYAR / ISOLIR / PSB; grafik income vs pengeluaran; monitoring per MikroTik.

### 6.2 MikroTik / Server / Tunnel / IP Publik
- Multi-server: `routeros_list` (nama, IP, port Winbox/API/FTP, user), `save_routeros/0`, `test_routeros`, `open_mikrotik`, `delete_mikrotik`, `pindah_server_device_*`, `set_session_id_mikrotik`, `connect/disconnect_mikrotik_server`, `new_server`, `save_server`, `update_server`, `delete_server`.
- Wizard menulis ke MikroTik via API: create secret PPP, clear port, buka port public/local (winbox/API/FTP), setup **L2TP global** atau **WireGuard ROS7**, pilih `server_tunnel`, generate script ROS6/ROS7; payload: `nama_server`, `remote_address`, `id_server`, `ip_address`, `public_*_port`, `local_*_port`, `username`, `password`, `versi_ros`, `public_key_wg`, `server_tunnel`, `tipe_ip_public`.
- IP Publik: `ip_public`, `ip_public/form_input`; dipakai wizard untuk membuat IP publik via remote address/PPP.

### 6.3 OLT
- Wizard `save_olt/add` berjalan per-step (JSON progres): `check_l2tp_global` → `update_firewall_mikrotik_mitra` → `update_secret` → `clear_port_olt` → `search_port_olt` → `create_port_telnet_server` / `create_port_web_server` → … → `finishing_save_olt`; ada resume `save_olt('{current_proses}')` dan `cancel_save_olt`.
- Field OLT (form & payload JS): `nama_olt`, `type_olt` (**CARD**/**BUNDLE**), `model_olt`, `serial_number_olt`, `ip_local_olt`, `ip_public_olt`, `telnet_local/public`, `snmp_local/public`, `web_local/public/protocol`, `snmp_ro`, `snmp_rw`, `username`, `password`, `auto_save` (1–6 jam), `id_mikrotik`, `key_hsgq/value_hsgq` (iframe vendor HSGQ), `olt_software_version`, `timeid` (token NAT).
- Brand/model yang terlihat: **CARD** (ZTE-C300/C320, Huawei, HsGQ, dll.) dan **BUNDLE** (HiOSO, VSOL-GLOBAL, VSOL-JOLINK, C-DATA, GL-xx), dengan catatan kompatibilitas telnet/web per brand.
- **Sinkronisasi 3 jalur**: `Upload DAT` (`syncronize_dat`, file `.dat`), **DAT Telnet** via *Server Global* (`syncronize_telnet`) atau *MikroTik Local* (`dat_telnet_mikrotik_info`), dan **Sync Cerdas** (step-by-step); pendukung: `syncronize`, `olt/field_data`, `olt/stale_list`.
- Tabel OLT menampilkan username/password/SNMP RO/RW — praktis untuk operasional, tetapi secret tampil apa adanya di UI.

### 6.4 GlobalACS (TR-069)
- ACS URL terpusat `http://acsmonitor.global-smartapp.com:321` + kredensial ACS global untuk diisikan ke modem (di laporan ini di-redaksi).
- Daftar ONT server-side (`acsGetDataPaginated`, `acsGetData?noSession=yes`), aksi **Summon/Refresh**, **Reboot**, **Reset Factory**, **Delete** (`acs/api/reboot|resetFactory|delete`, autentikasi **Bearer token** ke layanan ACS).
- Pencarian SN/MAC minimal 6 karakter; prefix internal `bysn:` / `bymac:`; hasil dapat langsung **ditambahkan/di-bind ke pelanggan** (`id_client`).
- Tutorial langkah: cek koneksi, profil VLAN, **scanning daftar PON interface (SNMP)**, fetch daftar ONU, dsb.

### 6.5 Pelanggan (Data Client) — inti billing
- Kolom tabel: Nama, Whatsapp, **ID Billing**, **Tgl PSB**, **Tgl Isolir**, **Last Payment**, ID Pelanggan, Server, Username, Password, Paket, Berlangganan, Collector, **Biaya**, Status, **Hutang**, **Prorata/Keterlambatan**, Alamat, Tipe (PPPoE/Hotspot/IP Static), **OLT, ONU, ODP**, NIK, Email, **Status Bayar (range H-7)**, Action, Wilayah, Teknisi, Sales, Catatan.
- Tab: Activated / Expired / **Hutang** / Pasang Baru / Incompleted / **Tagout** / Berhenti Langganan.
- Filter: server, collector, tipe pelanggan, expired (per tanggal + jumlah client), status pembayaran (Bayar Hari Ini / Bayar Bulan Ini / Sudah Bayar / Belum Bayar), lama expired (1–12 bulan).
- Aksi/endpoint: `list_invoice`, `confirm_payment_pppoe` (perpanjang `perpanjangan_sampai`), `cancel_payment_pppoe`, `input_income_pppoe` (bayar “other”: tanggal/jumlah/keterangan), `confirm_direct_pppoe_step`, `restart_koneksi_pppoe`, `check_koneksi_pppoe`, `setuptoisolir_pppoe|ipstatic`, `tagout/untag_billing_*`, `get_expired_dates`, `cek_password_client`, `simpan_set_password_client`, `get/simpan_wifi_data_client`, `bulk_delete_clients`, `save_user/0`, `update_id_pel`, dan **massal bayar** → `job_id` + `item_ids` → `router_job/*` (opsi notifikasi & restart modem).
- Modal **Konfirmasi Pembayaran**: review Order ID, tanggal, client, via transfer, bukti transfer, biaya tambahan, diskon, PPN, nominal sebelum approve; endpoint `save_wa_konfirmasi_manual`, `sending_wa_gateway_background[/stop]`, `getStatusSendingWA`.
- Modal **Prorata/Keterlambatan**: Harga Paket, PPN, **Denda**, Total; kolom `ProrataKeterlambatan` — perhitungan prorata + denda dilakukan server-side saat pembayaran.

### 6.6 Keuangan (Income, Pengeluaran, Hutang, Biaya Tambahan, Diskon, PPN)
- **Income** (`income/home`, `income/add/0`): ledger penerimaan (ref invoice, collector, tgl bayar, tipe, tujuan/bank, server, tipe income, client, alamat, username, profile, jumlah); hapus, **export Excel**, **print PDF**, **bulk download invoice**.
- **Pengeluaran** (`pengeluaran/home`): kategori + sisa budget (`pengeluaran/kategori_*`), input by, server, nominal; export.
- **Hutang Piutang** (`hutang_piutang`): hutang manual per client (tanggal, jumlah, jatuh tempo, status, tanggal bayar); pelunasan (`hutang_lunas`), hapus (`delete_hutang`); terhubung ke kolom Hutang & modal LIST HUTANG di Data Client.
- **Biaya Tambahan Langganan**: biaya berulang per pelanggan (keterangan, nominal, tipe, status, terakhir ditagihkan); toggle aktif.
- **Diskon**: per pelanggan per periode, Rp/%; diterapkan saat pembayaran (modal konfirmasi punya field diskon & PPN).
- **PPN**: setelan global mitra (`/account`, field `ppn`) + per pelanggan (`ppn`, `ppn-massal`); laporan pajak menghitung PPN amount per transaksi.
- **Prabayar/Pascabayar**, **Tanggal Tutup Buku 1–28**, **Hari Belum Bayar 1–90 (default 7; hanya tampilan/filter, tidak memengaruhi toleransi isolir)**.

### 6.7 Payment Gateway & Pembayaran Online
- Pilihan: **Global Moneta (Recommended)**, **Midtrans**, **No Using Payment Gateway**.
- Konfigurasi `save_pg?registrasi=`: `jenis_bank`, `nomor_rekening`, `nama_rekening`, `fee_admin_by` (100% Mitra / 100% Pelanggan / 50-50), `pin_wd`, `additional_fee_admin`, `biaya_penanganan`, `setting_title_invoice`, `setting_footer_invoice`, `show_in_invoice`, `status` (Sandbox/Production), kredensial `id_merchant_sand|prod`, `client_key_sand|prod`, `server_key_sand|prod`.
- Alur: pilih gateway → pelanggan bayar → konfirmasi (manual/dashboard) → income tercatat; halaman **Pembayaran** (`saldo/isi_saldo`) untuk verifikasi bukti manual.
- Payment method manual (`list_payment`): nama payment, nomor, A/N, gambar QRIS, status aktif; penerima konfirmasi manual.

### 6.8 PSB & Sales
- `planning_psb` + `planning_psb/add`: `name`, `phone_number`, `email`, `nik`, `alamat`, `latitude/longitude`, `foto_rumah`, `odp`, `paket_psb`, `planning_psb`, `wifi_name`, `wifi_password`; status pipeline dengan aksi **Update Status**, **Selesaikan PSB**, Foto Rumah, Detail Client.
- Pengaturan: `psb_setting/get_settings|save_setting|save_paket|delete_paket|save_syarat` (paket PSB: nama, kecepatan, harga, deskripsi, urutan, aktif; syarat; WA default; karyawan).
- Sales: `sales/ajax_list_client/pending`, `sales/set_id_pelanggan`, `sales/update_planning`, `sales/edit_client`, `sales/cancel_client`, `sales/delete_client`, `sales/report` (planning vs actual PSB).
- Registrasi publik per mitra: `daftar/{token}` (saat riset halaman publik maintenance).

### 6.9 Kolektor, Agen Hotspot, Absensi
- **Collector**: `collector_list_user` (total setor, sisa di tangan, status, last login) + `setoran_collector` (terkumpul, disetor, sisa, dikonfirmasi oleh).
- **Agent Hotspot**: total income, komisi, voucher belum/sudah terpakai, hutang voucher; form agent (profil, limit hutang voucher, prefix voucher, collector penagih); `bagikan`, `konfirmasi_voucher_agent`, `tagihan_agent`, `setoran_agent_hotspot`, export/upload template/tutorial.
- **Absensi**: `absensi` (tidak tersedia untuk akun ini), `absensi/rekap` (selfie, lokasi, device, masuk/pulang, status; export CSV), `absensi/setting` (kantor, mode keamanan); endpoint check-in/out, izin/sakit, approve/hapus.

### 6.10 Inventory
- Master barang: tipe, nama, satuan, harga satuan, min stok, sisa stok, nilai stok (`inventory/items`).
- Barang datang/keluar (`add_incoming/outgoing`, edit), stock opname (`add_opname`, stok sistem vs fisik vs selisih), export history.
- Modem/ONT: serial, MAC, garansi, status **In Stock/Assigned/Damaged/Lost**, relasi layanan/client/expired/pembayaran; `assign_modem` (PPPoE/Hotspot/IP Static), `return_modem`.

### 6.11 GIS/Topologi & Kalkulator
- `odp`, `odc`, `joint_closure`, `topologi`, `server`, `mapping_all_point` (HERE Maps) + import KMZ; relasi parent-child ODC/ODP (`connect_child_*`, `disconnect_child_*`, `get_*_children`), client ke ODP.
- Topologi menyimpan tipe (tiang/ODC/ODP/server), link-to, jarak, kelebihan kabel, slack, joint closure, rasio, **estimasi redaman**.
- Kalkulator: splitter ratio, jalur percabangan, 8 ODP ratio (konektor, splicing, kabel, ODP, rasio, output, **sisa laser**).

### 6.12 Notifikasi, WA/Telegram, Banner, APK
- **Notification Apps**: FCM push; target `all/odp/server/expired/tipe_user/pppoe_offline/custom`, pesan ≤500 char, gambar/action URL, **jadwal kirim otomatis per jam**, histori terkirim/gagal, template per tipe pesan, device token.
- **WhatsApp**: file/session device (`whatsapp/get-session`, `whatsapp/start-session`), pengiriman background + stop, status (`getStatusSendingWA`), dipakai untuk invoice/reminder/overdue/PSB/aktif/suspend/konfirmasi manual; modul **gated**.
- **Telegram**: token bot + aktivasi webhook (`activation_webhook_telegram`), notifikasi teknisi (ticket).
- **Billing banner** per server MikroTik (HTML body, tampil pada halaman billing/isolir); **APK client** distribusi app pelanggan.

### 6.13 Laporan
- **Summary**: klien expired + hari terlambat, stok inventory, tren.
- **Penjualan**: transaksi per periode; ringkasan per tipe koneksi/server/collector; export.
- **PSB**: laporan PSB/petugas/server/tipe user/invoice.
- **Tax**: PPN per transaksi + total per server/collector.
- **Sales**: planning vs actual PSB per sales.
- **Log Activity**: audit trail aplikasi + export.

---

## 7. Alur End-to-End Real ISP (direkonstruksi dari UI, form & endpoint)

1. **Onboarding mitra**: daftar di `/auth/registration` → login owner → isi profil mitra (PPN, Prabayar/Pascabayar, tutup buku) → tambah **Server** → tambah **MikroTik** (wizard tunnel L2TP/WireGuard + secret PPP + port publik) → tambah **OLT** (wizard tunnel/firewall/port + auto-save) → sync OLT (DAT/Telnet/Sync Cerdas) → tambah **ODC/ODP/JC/tiang** + topologi.
2. **Produk & billing**: set paket PSB/kecepatan/harga; atur prorata, diskon, PPN; metode pembayaran manual (bank/QRIS) dan/atau payment gateway (Moneta/Midtrans); set halaman isolir + banner.
3. **PSB**: registrasi (publik `daftar/{token}` atau input admin) → planning PSB (jadwal, sales, ODP) → assign teknisi → selesaikan PSB → pelanggan aktif (PPPoE/Hotspot/IP Static), OLT/ONU/ODP terhubung, inventory modem ter-assign, tagihan pertama.
4. **Operasional harian**: monitoring MikroTik/OLT/ACS, remote device, ticket teknisi (Telegram), pembayaran manual/gateway, mass payment, reminder WA/FCM, isolir otomatis (tagout), restore saat bayar (`confirm_payment_pppoe`/`untag`), collector menagih & setor, agent jual voucher & setor.
5. **Siklus bulanan**: tutup buku (tanggal 1–28) → tagihan/expired per pelanggan → reminder H-7/H-3/H-1 (status “Belum Bayar” mengikuti setelan hari) → pembayaran (diskon/PPN/denda/prorata) → perpanjangan expired → laporan penjualan/PSB/pajak/sales + income/expense + log audit.
6. **Pengembangan bisnis**: modul add-on (WA gateway, absensi, dst.) di-upgrade via `/add_ons`.

---

## 8. Pola Teknis yang Relevan untuk Polyglot

| Pola | Detail Global SmartApp | Implikasi untuk Polyglot |
|---|---|---|
| **Expired-driven billing** | Pembayaran menambah `perpanjangan_sampai`; denda/prorata dihitung saat bayar; invoice melekat pada pelanggan | Polyglot sudah invoice-driven + `paid_through`; pertimbangkan mode **PRABAYAR/PASCABAYAR** + tanggal tutup buku (`isp.billing_mode`, `isp.closing_date`) |
| **Biaya tambahan berulang** | `biaya_tambahan_langganan` (nominal/tipe/status/terakhir ditagih) otomatis masuk tagihan | Belum ada → tambahkan domain `recurring_extra_fee` + item invoice otomatis |
| **Denda/prorata keterlambatan** | Modal pembayaran menampilkan Harga/PPN/Denda/Total; kolom `ProrataKeterlambatan` | Tambahkan `late_fee` (flat/percent, grace) pada domain invoice/usecase pembayaran |
| **Hutang-piutang manual** | Hutang non-tagihan per pelanggan + pelunasan | Tambahkan tabel `customer_debts` + pembayaran hutang (bukan payment invoice) |
| **Mass payment + job** | Bayar massal → `job_id`+`item_ids` → `router_job/*`; progress persist antar halaman | Polyglot sudah `router_jobs`; tambahkan endpoint mass-pay + job monitor UI |
| **Payment gateway aggregator** | Global Moneta (bank/retail/e-wallet/QRIS) + fee bearer + withdrawal PIN | Pertimbangkan adapter agregator (selain Tripay/Midtrans/Xendit direct) + settlement/pencairan |
| **WA gateway device** | Sesi WA sendiri (QR), kirim background, status kirim; gated add-on | Polyglot pakai WA gateway existing; pertimbangkan device-session + background sender status |
| **Notifikasi pintar** | FCM push + jadwal per jam + segmentasi (ODP/server/expired/user type/PPPoE offline/custom) | Tambahkan segmentasi & penjadwalan notifikasi |
| **Wizard provisioning** | OLT/MikroTik onboarding bertahap, progress + resume/retry + token idempotensi NAT | Polyglot punya tunnel/telnet jobs; bisa ditambah progress streaming/step resume |
| **DAT sync 3 jalur** | Upload DAT / Telnet server global / Telnet MikroTik local / Sync Cerdas | Sebelumnya sengaja di luar scope; kandidat fitur opsional fase lanjutan |
| **Collector & agent** | Setoran & sisa di tangan, komisi agent, hutang voucher | Peluang modul keuangan lapangan (multi-role) |
| **Inventory/ONT** | Stok barang + modem/ONT lifecycle (assign/return/warranty) | Belum ada; berguna untuk PSB + aset |

---

## 9. Keamanan & Kepatuhan — Observasi (bukan uji penetrasi)

1. **2FA bisa ditunda** (“Ingatkan Nanti”) atau ditolak dengan pernyataan; pada akun riset, dashboard dapat diakses tanpa 2FA aktif — postur lemah untuk panel yang menyimpan kredensial OLT/router dan data pelanggan.
2. **Kredensial OLT (username/password/SNMP RO/RW) ditampilkan di tabel UI**; ACS memakai kredensial global yang ditampilkan untuk diisikan ke semua modem — model keamanan terpusat/berbagi.
3. **Header keamanan**: dari respons yang diamati hanya `cache-control: no-store`; tidak terlihat HSTS/CSP/X-Frame/X-Content-Type-Options; proteksi utama kemungkinan di Cloudflare/WAF.
4. **CSRF**: hidden `ci_csrf_token` ada, tetapi nilainya kosong pada halaman yang diamati; perlu verifikasi apakah diisi JS atau CSRF nonaktif.
5. **Sesi** `ci_session` HttpOnly+Secure ±2 jam; logout tersedia.
6. **Otorisasi**: pembatasan modul per role/permission diterapkan (403 + menu tersembunyi); **tidak** diuji cross-tenant (tidak etis/boleh dilakukan pada akun ini).
7. **Data sensitif**: hanya data tenant akun ini yang diambil; tidak ada data mitra lain yang diakses.

---

## 10. Perbandingan dengan Polyglot & Rekomendasi Prioritas

| Kapabilitas Global SmartApp | Status Polyglot (repo) | Rekomendasi |
|---|---|---|
| Invoice + prorata + diskon + PPN per paket | Sudah (billing v2) | Pertahankan; tambah **PPN global mitra** + per pelanggan |
| Prabayar/Pascabayar + tutup buku + hari belum bayar | Belum | **P1**: setelan `isp.billing_mode`, `isp.closing_date`, `isp.unpaid_days_display` |
| Biaya tambahan langganan berulang | Belum | **P1**: domain + auto-item saat generate invoice |
| Denda keterlambatan/prorata denda | Belum (hanya prorata awal) | **P1**: `late_fee` flat/percent + grace + preview kasir/portal |
| Hutang-piutang manual + pelunasan | Belum | **P2**: modul debts + pembayaran hutang |
| Mass payment + job progress | Job queue ada; belum mass payment | **P1**: endpoint mass-pay + job monitor UI |
| Payment gateway aggregator + fee bearer + WD | Direct Tripay/Midtrans/Xendit + fee bearer sudah ada | **P2**: adapter agregator (Moneta-like), settlement/withdrawal |
| WA gateway device + background sender + status | WA sender ada; gateway device belum | **P1/P2**: status kirim & retry background |
| FCM push + jadwal + segmentasi audiens | Notifikasi WA/email dasar | **P2**: FCM + segmentasi + schedule |
| Collector & setoran, agent hotspot & voucher | Belum | **P2** (pasar multi-role): collector settlement, agent voucher/komisi |
| Absensi (selfie+GPS) & permission granular | Belum | **P3** HR ops; permission granular bisa diadopsi lebih dulu |
| Inventory barang & modem/ONT lifecycle | Belum | **P2**: stok + assign ONT ke pelanggan (sinergi OLT/ACS) |
| ODC/ODP/JC/topologi + kalkulator redaman | Sengaja ditunda (di luar scope sebelumnya) | **P2**: modul FTTH/GIS terpisah |
| OLT wizard tunnel + 3 mode sync DAT | Driver OLT sudah; wizard belum | **P2** wizard; DAT sync kandidat fase lanjut |
| ACS summon/reboot/reset/bind | Sudah (OntService + GenieACS/OLT) | Pertahankan; tambah **summon/refresh** & bind UI |
| Billing banner & halaman isolir editable | Halaman isolir server-side ada; banner belum | **P2**: editor banner per server/device |
| Telegram bot + APK white-label + add-on gating | Belum | **P3** |
| Laporan penjualan/PSB/sales/tax + export PDF/Excel | Sebagian (finance report, aging, gateway recon, PSB stats) | **P1**: filter periode/server/collector + export Excel/PDF + laporan PPN |

---

## 11. Kesimpulan

Global SmartApp adalah **monolit PHP operasional ISP** dengan cakupan jauh lebih luas daripada “billing” biasa: jaringan (MikroTik/OLT/ACS/GIS), keuangan (billing, gateway, kas, utang, biaya tambahan), SDM lapangan (collector/agen/absensi), inventory, PSB, notifikasi multi-kanal, dan laporan pajak/sales — dengan **subscription add-on gating**.

Kekuatan yang layak ditiru Polyglot: **model masa aktif/expired yang sederhana untuk pasar RT/RW Net**, **biaya tambahan berulang + denda/prorata keterlambatan**, **collector/agent settlement**, **wizard provisioning bertahap dengan resume**, **mass operation berbasis job**, dan **segmentasi notifikasi**. Risiko yang terlihat: panel keamanan (2FA opsional/postponable, kredensial device tampil di UI, header keamanan minim), dan konsistensi UX antar paket add-on.

Untuk Polyglot, prioritas paling bernilai jangka pendek: **P1 — tutup buku/mode bayar, biaya tambahan berulang, denda keterlambatan, mass payment + job monitor, laporan PPN/penjualan + export**. Modul lapangan (collector/agent/inventory/GIS) sebaiknya menjadi roadmap terpisah agar fondasi billing tetap kuat.

---

## Lampiran C — Metodologi & Jejak

1. Recon publik: `GET /` (halaman login client), `GET /auth`, header server (`PHP/7.4.33`, Cloudflare).
2. Login terautentikasi `POST /auth` dengan akun mitra → redirect `/two_factor_setup`.
3. Setup 2FA TOTP terdeteksi; dipilih **“Ingatkan Nanti”** (tanpa mengubah keamanan akun) → dashboard `/admin`.
4. Crawl menu: 47 item sidebar diparse → 70 path; 69 halaman diambil **GET** dengan jeda 0,6 detik; disimpan sementara di `/tmp` (tidak dimasukkan ke repo karena memuat data akun).
5. Ekstraksi: form/field, tabel/kolom, modal, select options, inline JS, endpoint AJAX, variabel konfigurasi.
6. Fetch form read-only: `user_mng/add/0`, `planning_psb/add`, `income/add/0`, `list_payment/form_payment/0`, `ticket/add/0`, `agent_hotspot/add/0`, `ip_public/form_input`.
7. Logout `GET /auth/logout`; cookie & secret TOTP dihapus dari mesin riset.
8. **Tidak dilakukan**: aksi tulis/approve/delete, uji penetrasi, akses data mitra lain, eksploitasi celah, atau penyimpanan kredensial.

---

## Lampiran A — Sitemap Hasil Crawl (halaman admin)

| # | Path | HTTP | Judul Halaman |
|---|---|---|---|
| 1 | `/absensi` | 200 | Absensi |
| 2 | `/absensi/rekap` | 200 | Rekap Absensi |
| 3 | `/absensi/setting` | 200 | Pengaturan |
| 4 | `/account` | 200 | My Profile |
| 5 | `/acs` | 200 | GlobalACS |
| 6 | `/add_ons` | 200 | Akses Terbatas |
| 7 | `/admin` | 200 | Dashboard GLOBAL |
| 8 | `/admin?ip_public=no` | 200 | Dashboard GLOBAL |
| 9 | `/agent_hotspot` | 200 | Agent Hotspot |
| 10 | `/agent_hotspot/bagikan` | 200 | Bagikan Voucher ke Agent |
| 11 | `/apk` | 200 | APK CLIENT |
| 12 | `/biaya_tambahan_langganan` | 200 | Biaya Tambahan Langganan |
| 13 | `/billing_banner` | 200 | Billing Banner |
| 14 | `/burst_limit` | 200 | Burst Limit |
| 15 | `/calc_ratio` | 200 | Calculator Ratio |
| 16 | `/calc_ratio/jalur` | 200 | Jalur Percabangan |
| 17 | `/calc_ratio/odp` | 200 | 8 ODP Ratio |
| 18 | `/calc_ratio/splitter` | 200 | Splitter Ratio |
| 19 | `/collector_list_user` | 200 | List Collector |
| 20 | `/discount` | 200 | Discount |
| 21 | `/documentation/home` | 200 | Documentation |
| 22 | `/home_isolir` | 200 | Set Page Isolir |
| 23 | `/hutang_piutang` | 200 | Hutang Pituang |
| 24 | `/income/home` | 200 | Income  All Tipe |
| 25 | `/inventory/incoming` | 200 | Barang Datang |
| 26 | `/inventory/items` | 200 | Inventory Items |
| 27 | `/inventory/modems` | 200 | Inventory Modems |
| 28 | `/inventory/opname` | 200 | Stock Opname |
| 29 | `/inventory/outgoing` | 200 | Barang Keluar |
| 30 | `/inventory/tutorial` | 200 | Tutorial Inventory |
| 31 | `/ip_public` | 200 | IP PUBLIC |
| 32 | `/joint_closure` | 200 | Joint Closure |
| 33 | `/konfirmasi_voucher_agent` | 200 | KONFIRMASI VOUCHER AGENT |
| 34 | `/list_payment` | 200 | Payment Method |
| 35 | `/logsActivity` | 200 | Logs Activity Apps |
| 36 | `/mapping_all_point?first=yes` | 200 | Mapping All Point |
| 37 | `/notification` | 200 | Notification Apps |
| 38 | `/odc` | 200 | ODC List |
| 39 | `/odp` | 200 | ODP List |
| 40 | `/olt/form_olt/add` | 200 | Tambah OLT |
| 41 | `/olt/setting` | 200 | OLT Setting |
| 42 | `/payment_gateway` | 200 | Payment Gateway |
| 43 | `/payment_gateway?registrasi=global` | 200 | Payment Gateway |
| 44 | `/pengeluaran/home` | 200 | Pengeluaran |
| 45 | `/planning_psb` | 200 | PLANNING PSB |
| 46 | `/prorata` | 200 | Prorata |
| 47 | `/remote_device` | 200 | Remote Device |
| 48 | `/report_penjualan` | 200 | Report Penjualan |
| 49 | `/report_psb` | 200 | Report PSB |
| 50 | `/report_summary` | 200 | Report Summary |
| 51 | `/report_tax` | 200 | Report Tax |
| 52 | `/routeros/0` | 200 | Tambah Mikrotik |
| 53 | `/routeros_list` | 200 | Mikrotik |
| 54 | `/saldo/isi_saldo` | 200 | Pembayaran |
| 55 | `/sales/report` | 200 | REPORT SALES OCT 2026 |
| 56 | `/server` | 200 | Server List |
| 57 | `/setoran_agent_hotspot` | 200 | SETORAN AGENT HOTSPOT |
| 58 | `/setoran_collector` | 200 | SETORAN COLLECTOR |
| 59 | `/syncron_income_hotspot` | 200 | - |
| 60 | `/tagihan_agent` | 200 | Tagihan Agent Hotspot |
| 61 | `/telegram_bot/create` | 200 | Buat Telegram Bot |
| 62 | `/ticket` | 200 | TICKET |
| 63 | `/topologi` | 200 | Planing Topologi Jaringan |
| 64 | `/user_list` | 200 | Data Client |
| 65 | `/user_list?q=belum_bayar` | 200 | Data Client |
| 66 | `/user_list?q=isolir` | 200 | Data Client |
| 67 | `/user_list?q=psb` | 200 | Data Client |
| 68 | `/user_list?q=sudah_bayar` | 200 | Data Client |
| 69 | `/user_mng` | 200 | Karyawan |

*Total: 69 halaman admin (crawl GET, read-only, rate-limited).*

---

## Lampiran B — Inventaris Endpoint

Terdapat **273 endpoint unik** yang direferensikan di HTML/JS halaman admin (kategori: auth/account, router/server/tunnel, OLT, ACS, pelanggan, billing/payment, keuangan, PSB/sales, laporan, notifikasi, inventory, GIS, ticket, absensi/collector/agent). Daftar lengkap:

```text
/absensi
/absensi/admin_save
/absensi/approve
/absensi/check_in
/absensi/check_out
/absensi/hapus
/absensi/izin_sakit
/absensi/rekap
/absensi/rekap/dt_data
/absensi/rekap/export_csv
/absensi/setting
/absensi/setting/delete
/absensi/setting/get
/absensi/setting/mode_keamanan
/absensi/setting/save
/account
/acs
/activation_webhook_telegram
/add_ons
/add_ons/add
/add_topologi
/admin
/agent_hotspot
/agent_hotspot/add/0
/agent_hotspot/bagikan
/agent_hotspot/export_excel
/agent_hotspot/tutorial
/agent_hotspot/upload_template
/agent_hotspot_save/add/0
/api/load_penjualan
/api/load_psb
/api/load_tax
/apk
/biaya_tambahan_langganan
/billing_banner
/billing_banner_delete
/billing_banner_get_data
/billing_banner_get_detail
/billing_banner_preview
/billing_banner_save
/bulk_delete_clients
/burst_limit
/calc_ratio
/calc_ratio/jalur
/calc_ratio/odp
/calc_ratio/splitter
/cancel_payment_pppoe
/cancel_save_olt
/cancel_save_routeros/0
/cek_password_client
/change_logo_owner
/change_password
/change_photo_profile
/check_koneksi_pppoe
/check_ping
/checking_page
/checking_page_olt
/checking_password
/collector_list_user
/confirm_direct_pppoe_step
/confirm_payment_pppoe
/connect_child_odc
/connect_child_odp
/connect_client_to_odp
/daftar/{token_mitra}
/dashboard_client
/data_interface_hotspot
/delete_biaya_tambahan_langganan
/delete_hutang
/delete_income
/delete_joint_closure
/delete_mikrotik
/delete_odc
/delete_odp
/delete_olt
/delete_pengeluaran
/delete_server
/detail_mikrotik
/disconnect_child_odc
/disconnect_child_odp
/disconnect_client_from_odp
/disconnect_mikrotik_server
/discount
/discount/add/0
/documentation/home
/download_isolir
/editing_account
/endpoint
/fpass
/getStatusSendingWA
/get_available_children_serverside
/get_available_hotspot_serverside
/get_available_ipstatic_serverside
/get_available_odp_children_serverside
/get_available_pppoe_serverside
/get_clients_odp
/get_connected_children
/get_connected_odp_children
/get_expired_dates
/get_info_dashboard
/get_joint_closure_map_data
/get_list_servers
/get_odc_by_server
/get_odp_by_odc
/get_odp_points
/get_traffic
/get_traffic_income
/get_traffic_pengeluaran
/get_wifi_data_client
/getlog
/help
/home_isolir
/hotspot_printvoucher
/hutang_lunas
/hutang_piutang
/income/add/0
/income/bulk_download_invoice
/income/download_excel
/income/home
/income/print_pdf
/income/serverside
/input_income_pppoe
/inventory/add_incoming
/inventory/add_item
/inventory/add_modem
/inventory/add_opname
/inventory/add_outgoing
/inventory/assign_modem
/inventory/edit_incoming
/inventory/edit_modem
/inventory/edit_opname
/inventory/edit_outgoing
/inventory/export_history/Incoming
/inventory/export_history/Opname
/inventory/export_history/Outgoing
/inventory/export_items
/inventory/get_clients_by_type
/inventory/incoming
/inventory/items
/inventory/modems
/inventory/opname
/inventory/outgoing
/inventory/return_modem
/inventory/tutorial
/ip_public
/ip_public/form_input
/joint_closure
/konfirmasi_voucher_agent
/list_payment
/list_payment/form_payment/0
/login_client
/logsActivity
/logsActivity/data
/logsActivity/download
/manifest
/mapping_all_point
/mikrotik/setting_hotspot
/new_server
/notification
/odc
/odp
/olt/field_data
/olt/form_olt/add
/olt/setting
/olt/stale_list
/open_mikrotik
/payed_md/{token_mitra}
/payment_gateway
/pengeluaran/download_excel
/pengeluaran/home
/pengeluaran/kategori_delete
/pengeluaran/kategori_list
/pengeluaran/kategori_save
/pengeluaran/kategori_sisa_budget
/pengeluaran/serverside
/pindah_server_device_check
/pindah_server_device_exec
/pindah_server_device_info
/planning_psb
/planning_psb/add
/prorata
/psb_setting/delete_paket
/psb_setting/get_settings
/psb_setting/save_paket
/psb_setting/save_setting
/psb_setting/save_syarat
/psb_update_status
/remote_device
/report_penjualan
/report_psb
/report_summary
/report_tax
/resend_email_activation
/restart_koneksi_pppoe
/router_job/ack
/router_job/status
/router_job/step
/router_job/stop
/routeros/0
/routeros_list
/saldo/isi_saldo
/sales/ajax_list_client/pending
/sales/cancel_client
/sales/delete_client
/sales/edit_client
/sales/report
/sales/save_user_psb/0
/sales/set_id_pelanggan
/sales/update_planning
/save_biaya_tambahan_langganan
/save_hutang_piutang
/save_income/add/0
/save_joint_closure
/save_odc
/save_olt/add
/save_payment/0
/save_pengeluaran
/save_pg
/save_routeros/0
/save_serial_olt
/save_server
/save_ticket/add/0
/save_user/0
/save_user_mng/add/0
/save_wa_konfirmasi_manual
/sending_wa_gateway_background
/sending_wa_gateway_background/stop
/server
/service-worker
/set_default_password_client
/set_session_id_mikrotik
/set_session_id_olt
/setoran_agent_hotspot
/setoran_collector
/setuptoisolir_ipstatic
/setuptoisolir_pppoe
/simpan_odp
/simpan_set_password_client
/simpan_wifi_data_client
/syncron_income_hotspot
/syncronize
/syncronize_dat
/tagihan_agent
/tagout_billing_hotspot_user
/tagout_billing_ip_static
/tagout_billing_pppoe
/telegram_bot/create
/telegram_bot/save
/template_voucher
/test_routeros
/ticket
/ticket/add/0
/ticket_data
/ticket_delete
/ticket_update_status
/toggle_biaya_tambahan_langganan
/topologi
/two_factor_setup
/untag_billing_hotspot_user
/untag_billing_ip_static
/untag_billing_pppoe
/update_hari_belum_bayar
/update_id_pel
/update_joint_closure
/update_odc
/update_server
/user/get_data_user_ajax
/user_list
/user_mng
/user_mng/add/0
/whatsapp/get-session
/whatsapp/start-session
/whatsapp_gateway
```

> Catatan: metode HTTP dan parameter tidak diverifikasi satu per satu (banyak dipanggil via AJAX POST); daftar ini adalah referensi navigasi/riset, bukan dokumentasi API.
