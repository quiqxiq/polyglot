# Dokumentasi & Analisis Reverse Engineering 6 Modul Inti Global SmartApp
**Spesifikasi Teknis Billing & Network Management RT/RW Net & FTTH Cloud**  
*URL Target*: [https://app.global-smartapp.com](https://app.global-smartapp.com)  
*Akun Audit*: `quiqxiq@gmail.com` (Nama: Iqmam | ID Mitra: `6b397377415758324465`)  
*Status Langganan*: Trial Paket LITE II (Aktif s/d 12 Oktober 2026)

---

## Daftar Isi
1. [Arsitektur Sistem & Komunikasi Cloud-to-LAN](#1-arsitektur-sistem--komunikasi-cloud-to-lan)
2. [Status Akun & Reverse Engineering Add-on Remote Device](#2-status-akun--reverse-engineering-add-on-remote-device)
3. [Manajemen Perangkat Jaringan (Network Core: MikroTik, OLT, ACS TR-069)](#3-manajemen-perangkat-jaringan-network-core)
   - 3.1 [MikroTik RouterOS (ROS v6 vs ROS v7, WireGuard Script, Watchdog, RouterJob Worker)](#31-mikrotik-routeros)
   - 3.2 [OLT Management (HSGQ, VSOL, ZTE, Huawei, Telnet, SNMP, Auto-Save)](#32-olt-management)
   - 3.3 [GlobalACS - TR-069 Auto Configuration Server](#33-globalacs---tr-069-auto-configuration-server)
   - 3.4 [Remote Device Reverse Proxy](#34-remote-device-reverse-proxy)
4. [Perencanaan Infrastruktur FTTH & Kalkulator Redaman Optik](#4-perencanaan-infrastruktur-ftth--kalkulator-redaman-optik)
   - 4.1 [Hirarki Node FTTH (Server, ODC, Joint Closure, ODP)](#41-hirarki-node-ftth)
   - 4.2 [Pemetaan GIS & Port Allocation (`/mapping_all_point`)](#42-pemetaan-gis--port-allocation)
   - 4.3 [Desain Topologi Kabel Fiber (`/topologi`)](#43-desain-topologi-kabel-fiber)
   - 4.4 [Bedah Algoritma & Rumus Matematika Kalkulator Rasio (`calc_ratio`)](#44-bedah-algoritma--rumus-matematika-kalkulator-rasio)
5. [Manajemen Pelanggan, Layanan & Sistem Isolir Otomatis](#5-manajemen-pelanggan-layanan--sistem-isolir-otomatis)
   - 5.1 [Layanan PPPoE (Secret, Profile, Binding ONU, Sync DAT)](#51-layanan-pppoe)
   - 5.2 [Layanan Hotspot Member & Voucher (IP Binding, Rate Limit, Expiration)](#52-layanan-hotspot-member--voucher)
   - 5.3 [Layanan IP Static (Simple Queue, Burst Calculation, Priority, Bucket Size)](#53-layanan-ip-static)
   - 5.4 [Deep Dive Sistem Isolir Otomatis (WebProxy, NAT DST-NAT, Bypass Rule, Error Template)](#54-deep-dive-sistem-isolir-otomatis)
   - 5.5 [Workflow Planning PSB (Pasang Baru Mandiri & Lifecycle Pemasangan)](#55-workflow-planning-psb)
   - 5.6 [Prorata Billing & Diskon](#56-prorata-billing--diskon)
6. [Sistem Billing, Keuangan & Payment Gateway](#6-sistem-billing-keuangan--payment-gateway)
   - 6.1 [Payment Gateway Global Moneta (Auto Account Validation & Instant Un-isolir)](#61-payment-gateway-global-moneta)
   - 6.2 [Integrasi Midtrans (Sandbox vs Production)](#62-integrasi-midtrans)
   - 6.3 [Sistem Pembayaran Manual, Upload Bukti Transfer & Approval Approval](#63-sistem-pembayaran-manual-upload-bukti-transfer--approval)
   - 6.4 [Struktur Pembukuan Arus Kas (Income, Pengeluaran, Hutang-Piutang)](#64-struktur-pembukuan-arus-kas)

---

## 1. Arsitektur Sistem & Komunikasi Cloud-to-LAN

Global SmartApp dibangun dengan arsitektur **Hybrid Cloud Network Management**:
- **Backend Core**: CodeIgniter PHP Framework dengan MySQL Database pada cloud hosting.
- **Frontend Layer**: Dashboard berbasis tema modern SB Admin 2 / Skydash dengan AJAX Datatables, Select2, dan SweetAlert2.
- **Tantangan Utama Jaringan RT/RW Net**: Sebagian besar router MikroTik dan OLT milik ISP lokal berada di balik koneksi NAT ISP upstream (CGNAT / IP Privat Dinamis) tanpa IP Publik Statis.
- **Solusi Arsitektur**: Global SmartApp menyediakan **Cloud Tunnel Hub Gateway** terpusat (`mainremote.global-smartapp.com`). Router MikroTik lokal bertindak sebagai *Tunnel Client* yang membangun koneksi keluar (*outbound*) permanen ke cloud hub via WireGuard (port 13230) atau OpenVPN (port 1194). Melalui tunnel virtual ini, server cloud dapat mengirimkan perintah API MikroTik (port 8728) dan mengakses Web GUI / Telnet OLT secara dua arah.

```
+-----------------------------------------------------------------------------------+
|                           GLOBAL SMARTAPP CLOUD SERVER                            |
|                                                                                   |
|  [ Web Billing UI ]   [ RouterJob Worker ]   [ ACS TR-069 ]   [ Payment Webhook ] |
|  - Data Pelanggan     - Asynchronous Queue   - Port 321        - Global Moneta    |
|  - Hitung Redaman     - Step-by-Step Sync    - CWMP Protocol   - Midtrans         |
+-----------------------------------------+-----------------------------------------+
                                          |
                +-------------------------+-------------------------+
                | WireGuard Tunnel (UDP 13230) / OpenVPN           | TR-069 (HTTP/TCP 321)
                ▼                                                   ▼
+-------------------------------+                   +-------------------------------+
|     MIKROTIK ROUTEROS LOCAL   |                   |      MODEM / ONT PELANGGAN    |
| - Wireguard: --WG-TUNNEL01    |                   | - TR-069 Client               |
| - IP Tunnel: 178.50.90.8/32   |                   | - URL: acsmonitor...:321      |
| - Proxy Isolir: Port 3125     |                   | - RX Optical Power Telemetry  |
| - Scheduler Watchdog (1 min)  |                   | - WiFi SSID & Client Count    |
+---------------+---------------+                   +-------------------------------+
                | Local LAN / Telnet / SNMP
                ▼
+-------------------------------+
|           OLT LOCAL           |
| - HSGQ / VSOL / ZTE / Huawei  |
| - Telnet Port 23              |
| - SNMP Port 161 (RW / RO)     |
| - Auto-Save Config (1-6 Jam)  |
+-------------------------------+
```

---

## 2. Status Akun & Reverse Engineering Add-on Remote Device

### 2.1 Analisis Error 403 pada `/add_ons/add?r=remote_device`
Ketika mengakses URL `https://app.global-smartapp.com/add_ons/add?r=remote_device`:
- **HTTP Code**: 200 OK secara protokol HTTP, tetapi aplikasi menampilkan render template view `403 Akses Terbatas`.
- **Logika Middleware Backend**:
  ```php
  // Pseudocode logic otorisasi add_ons
  if ($user->subscription_status == 'trial' || $user->package_tier == 'trial_free') {
      return view('errors/access_limited', [
          'message' => 'Untuk menggunakan fitur ini, silakan upgrade akun Anda ke paket Berlangganan.'
      ]);
  }
  ```
- **Kesimpulan**: Add-on adalah modul penjualan kuota tambahan (misal: penambahan kuota router MikroTik, kuota OLT, kuota remote device unlimited). Akun dalam status masa percobaan (Trial LITE II) tidak diizinkan membuka formulir pembelian add-on sebelum menjadi pelanggan berbayar aktif.

### 2.2 Cara Kerja Fitur Remote Device Bawaan (`/remote_device`)
Akun saat ini mendapatkan jatah **10 Free Trial Client Remote**.

#### Mekanisme Reverse Proxy Remote Device:
1. **Pendaftaran Perangkat**:
   Pelanggan yang terdaftar dengan tipe modem ONU (misal ZTE F609, Huawei HG8245H, Fiberhome) memiliki IP lokal di MikroTik (misal IP WAN PPPoE: `10.10.10.55` atau IP LAN `192.168.1.1`).
2. **Reverse Port Mapping**:
   Ketika tombol **Action -> Open Remote** ditekan:
   - Server Cloud Global SmartApp mengirimkan instruksi via MikroTik API untuk membuat aturan temporary port-forwarding / SSH port-forwarding tunnel.
   - MikroTik lokal memetakan port 80 (Web GUI Modem) ke IP lokal tunnel (`178.50.90.8:port_acak`).
3. **Akses Browser**:
   Browser teknisi diarahkan ke URL reverse proxy (misal: `https://remote01.global-smartapp.com:port_acak/`) yang langsung menampilkan halaman login Web GUI modem pelanggan tanpa membutuhkan IP publik statis di rumah pelanggan maupun di router MikroTik.

---

## 3. Manajemen Perangkat Jaringan (Network Core)

### 3.1 MikroTik RouterOS

#### A. Komparasi Script Setup: ROS v6 vs ROS v7

##### 1. RouterOS v7 (Menggunakan WireGuard Native)
```routeros
# 1. Bersihkan interface tunnel lama
/interface ovpn-client remove [find name="--TUNNEL01-GLOBAL"];
/interface ovpn-client remove [find name="--TUNNEL02-GLOBAL"];
/interface ovpn-client remove [find name="--TUNNEL03-GLOBAL"];
/interface ovpn-client remove [find name="--REMOTE01-GLOBAL"];
/interface ovpn-client remove [find name="--REMOTE02-GLOBAL"];

# 2. Inisialisasi interface WireGuard jika belum ada
:if ([:len [/interface wireguard find name="--WG-TUNNEL01-GLOBAL"]] = 0) do={
    /interface wireguard add comment="---> (WG TUNNEL GLOBAL) Jangan mengganti nama dan menghapus interface dari billing" listen-port=34381 mtu=1420 name=--WG-TUNNEL01-GLOBAL;
    /ip address add address=178.50.90.8 interface=--WG-TUNNEL01-GLOBAL network=178.50.0.1;
} else={
    :log info "WG-TUNNEL01-GLOBAL sudah ada";
}

# 3. Binding Peer ke Cloud Hub Global SmartApp
/interface wireguard peers remove [find public-key="hypleh6DUTc73JZ/KtEamS8tofLJ4lL0n8oBEhVCQyc="];
/ip address remove [find address=178.50.90.8/32];
/interface wireguard peers add allowed-address=0.0.0.0/0 endpoint-address=mainremote.global-smartapp.com endpoint-port=13230 interface=--WG-TUNNEL01-GLOBAL persistent-keepalive=25s public-key="hypleh6DUTc73JZ/KtEamS8tofLJ4lL0n8oBEhVCQyc=";

# 4. Watchdog Scheduler (Self-Healing Connection)
/system scheduler remove [find name="--TUNNEL-CHECK-GLOBAL"];
/system scheduler add name="--TUNNEL-CHECK-GLOBAL" interval=1m on-event=":if ([/ping 178.50.0.1 count=5] = 0) do={/interface disable \"--WG-TUNNEL01-GLOBAL\"; :delay 5s; /interface enable \"--WG-TUNNEL01-GLOBAL\"; :log warning \"Tunnel RTO, interface reset done!\"}";
```

##### 2. RouterOS v6 (Menggunakan OpenVPN Client)
```routeros
/interface ovpn-client remove [find name="--TUNNEL01-GLOBAL"];
/interface ovpn-client add comment="---> Jangan mengganti nama dan menghapus interface dari billing" connect-to=tun01.global-smartapp.com name=--TUNNEL01-GLOBAL password=IQMAM3346-178.50.90.8 user=IQMAM3346-178.50.90.8 profile=default-encryption;
/system scheduler remove [find name="--TUNNEL-CHECK-GLOBAL"];
/system scheduler add name="--TUNNEL-CHECK-GLOBAL" interval=1m on-event=":if ([/ping 178.50.90.1 count=5] = 0) do={/interface disable \"--TUNNEL01-GLOBAL\"; :delay 5s; /interface enable \"--TUNNEL01-GLOBAL\"; :log warning \"Tunnel RTO, interface reset done!\"}";
```

#### B. RouterJob Asynchronous Queue Worker
Global SmartApp mengimplementasikan sistem antrean client-side di JavaScript untuk mengeksekusi operasi massal di MikroTik (seperti isolir ratusan pelanggan atau sinkronisasi data ribuan secret PPPoE):
- **Endpoint API Internal**:
  - `POST /router_job/status`: Memeriksa apakah ada proses antrean yang sedang berjalan.
  - `POST /router_job/step`: Menjalankan 1 batch operasi (misal 20 user) ke API MikroTik.
  - `POST /router_job/stop`: Membatalkan antrean operasi.
  - `POST /router_job/ack`: Mengirimkan konfirmasi bahwa batch selesai diproses.
- **Toleransi Kesalahan**:
  - `MAX_STEP_ERRORS = 5`: Menghentikan operasi jika server merespons error 5 kali beruntun.
  - `MAX_NET_ERRORS = 20`: Mengizinkan jeda sementara jika koneksi internet terputus sebelum beralih ke mode manual download file script `.rsc`.

---

### 3.2 OLT Management

#### A. Dukungan Perangkat & Protokol
Form `/olt/form_olt/add` mendukung varian OLT berbasis chassis (Card) dan kotak mandiri (Bundle):
- **Tipe Bundle**: C-DATA, GL-01, GL-02-128-CL, GL-02-128-VLU, GL-02-EGGV, GL-04, GL-08, HIOSO, HSGQ, VSOL-GLOBAL, VSOL-JOLINK.
- **Tipe Card (Chassis)**: HUAWEI (MA5608T/5683T), ZTE-C300, ZTE-C320.
- **Teknologi PON**: EPON & GPON (2 PON, 4 PON, 8 PON, 16 PON).

#### B. Parameter Koneksi & Konfigurasi Input:
1. `tipe-ip-public`: Pilihan `Gunakan IP Public Global` (via tunnel) atau `IP Public Sendiri`.
2. `name`: Label OLT di dashboard.
3. `remote-address`: ID MikroTik yang bertindak sebagai gateway LAN OLT.
4. `ip-local-olt`: IP management OLT pada jaringan lokal (misal: `192.168.10.100`).
5. `telnet-local`: Port Telnet lokal (default `23`).
6. `snmp-local`: Port SNMP lokal (default `161`).
7. `snmp-ro` & `snmp-rw`: Community string SNMP Read-Only dan Read-Write (contoh bawaan sistem: `1b4b0335` & `0f700267`).
8. `web-local`: Port Web Management OLT (default `80` HTTP / `443` HTTPS).
9. `auto-save`: Scheduler auto-save konfigurasi OLT ke NVRAM/Flash (pilihan: 1, 2, 3, 4, 5, atau 6 jam).
10. `username` & `password`: Kredensial login akun privileged/admin OLT.

#### C. Mekanisme Sinkronisasi Data OLT:
Sistem menyediakan 2 jalur sinkronisasi:
1. **Jalur Telnet Otomatis**: Server cloud memanggil perintah CLI OLT (seperti `show onu unconfigured`, `show gpon onu state`, `write all`).
2. **Jalur File Backup DAT**: OLT mengunggah file `startrun.dat` via FTP lokal MikroTik, lalu file tersebut di-upload ke sistem billing untuk membaca database ONU secara offline tanpa risiko sesi Telnet terputus.

---

### 3.3 GlobalACS - TR-069 Auto Configuration Server

#### A. Arsitektur ACS (`/acs`)
GlobalACS berfungsi sebagai sistem Auto Configuration Server berbasis standar broadband TR-069 / CWMP (CPE WAN Management Protocol):
- **Server Endpoint URL**: `http://acsmonitor.global-smartapp.com:321`
- **Konfigurasi pada Modem/ONT Pelanggan**:
  - ACS URL: `http://acsmonitor.global-smartapp.com:321`
  - Periodic Inform Interval: `300` detik (5 menit).

#### B. Parameter Telemetri yang Dibaca Real-Time dari Modem:
| Parameter | Keterangan Teknis |
|---|---|
| **Optical RX Power** | Kuat sinyal redaman laser optic yang diterima ONT (satuan: dBm, misal: `-21.4 dBm`). |
| **Temperature** | Suhu operasional internal chipset ONT (satuan: Celsius). |
| **SSID & Password** | Nama pemancar WiFi dan security key WPA2 yang aktif. |
| **Connected Devices** | Jumlah perangkat gadget (HP, laptop, Smart TV) yang sedang terhubung ke WiFi ONT. |
| **Serial Number (SN)** | Nomor seri fisik unik ONT (misal: `ZTEGC1234567` / `HWTC89ABCDEF`). |
| **Product Class & Soft Version**| Tipe hardware modem dan versi firmware sistem. |
| **IP WAN / TR-069** | IP yang didapatkan ONT untuk jalur internet dan jalur manajemen CWMP. |
| **Last Inform** | Waktu kontak terakhir antara modem dan server ACS. |

#### C. Tindakan Remote (Remote Actions):
1. **Reboot Modem**: Mengirimkan RPC method `Reboot` tanpa harus menyentuh tombol power fisik.
2. **Ganti SSID / Sandi WiFi**: Mengirimkan RPC method `SetParameterValues` untuk object `InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.PreSharedKey.1.KeyPassphrase`.
3. **Push Konfigurasi WAN PPPoE**: Mengatur username/password PPPoE langsung dari cloud.

---

### 3.4 Remote Device Reverse Proxy

- **Tujuan**: Membuka antarmuka grafis (Web GUI) modem ONT yang tidak memiliki TR-069 atau membutuhkan konfigurasi menu khusus (Port Forwarding, DMZ, Binding VLAN).
- **Alur Kerja**:
  1. Operator memilih pelanggan di `/remote_device` dan menekan **Open Remote**.
  2. Gateway Global SmartApp meneruskan paket HTTP via WireGuard tunnel ke MikroTik lokal.
  3. MikroTik merutekan trafik ke IP lokal ONT pelanggan.
  4. Web GUI modem tampil di tab browser teknisi dalam frame proxy terisolasi.

---

## 4. Perencanaan Infrastruktur FTTH & Kalkulator Redaman Optik

### 4.1 Hirarki Node FTTH
Global SmartApp memodelkan jaringan optik pasif (PON) dengan struktur berjenjang:
```
[ OLT Headend ]
      │ Feeder Cable (24 - 96 Core)
      ▼
   [ ODC ] (Optical Distribution Cabinet)
      │ Distribution Cable (12 - 24 Core)
      ▼
[ Joint Closure ] (Sambungan Kabel Tiang / Manhole)
      │ Distribution Sub-cable
      ▼
   [ ODP ] (Optical Distribution Point - 4 / 8 / 16 Port)
      │ Dropcore Cable 1 Core (Panjang 50 - 250 Meter)
      ▼
  [ Roset / OTP ]
      │ Patchcord SC/UPC
      ▼
[ ONT Pelanggan ]
```

### 4.2 Pemetaan GIS & Port Allocation (`/mapping_all_point`)
- Menggunakan peta berbasis peta satelit untuk memetakan koordinat latitude/longitude setiap titik ODP dan rumah pelanggan.
- **Port Allocation**: Setiap ODP mencatat nomor port yang telah terpakai (1 s/d 8 atau 1 s/d 16), nama pelanggan yang menancap di port tersebut, dan port kosong yang masih tersedia (*free capacity*).

---

### 4.4 Bedah Algoritma & Rumus Matematika Kalkulator Rasio (`calc_ratio`)

Berdasarkan hasil ekstraksi kode JavaScript pada `calc_ratio_splitter.html` dan `calc_ratio_odp.html`, berikut adalah rumus pasti yang digunakan sistem:

#### A. Rumus Akumulasi Total Loss Redaman:
$$\text{Total Loss (dB)} = \text{Loss Konektor} + \text{Loss Splicing} + \text{Loss Panjang Kabel} + \text{Loss ODP / Splitter}$$

Di mana:
1. **Loss Konektor Adapter**:
   $$\text{Loss}_{\text{konektor}} = -(\text{Jumlah Konektor} \times 0.30\text{ dB})$$
2. **Loss Sambungan Fusion Splicing**:
   $$\text{Loss}_{\text{splicing}} = -(\text{Jumlah Sambungan} \times 0.10\text{ dB})$$
3. **Loss Kabel Fiber Optic (Panjang Kabel)**:
   $$\text{Loss}_{\text{kabel}} = -(\text{Panjang Kabel dalam Meter} \times 0.00035\text{ dB})$$
   *(Ekuivalen dengan $0.35\text{ dB/km}$ pada panjang gelombang 1310/1490nm)*.
4. **Loss Splitter Simetris (ODP Box)**:
   - Splitter 1:2 = $-3.70\text{ dB}$
   - Splitter 1:4 = $-7.20\text{ dB}$
   - Splitter 1:8 = $-10.50\text{ dB}$
   - Splitter 1:16 = $-13.80\text{ dB}$

#### B. Tabel Lookup Rasio Asimetris (`kondisikeluaran(nilai)`):
Fungsi JavaScript `kondisikeluaran(nilai)` memetakan sisa redaman target ke persentase splitter rasio asimetris untuk jalur *cascade bus topology*:

| Nilai Redaman Output Target | Persentase Rasio Splitter | Nilai Redaman Output Target | Persentase Rasio Splitter |
|---|---|---|---|
| $< -18.69\text{ dB}$ | **1%** | $< -8.04\text{ dB}$ | **15%** |
| $< -16.31\text{ dB}$ | **2%** | $< -7.42\text{ dB}$ | **18%** |
| $< -14.80\text{ dB}$ | **3%** | $< -6.98\text{ dB}$ | **20%** |
| $< -13.69\text{ dB}$ | **4%** | $< -6.50\text{ dB}$ | **22%** |
| $< -12.81\text{ dB}$ | **5%** | $< -5.97\text{ dB}$ | **25%** |
| $< -12.08\text{ dB}$ | **6%** | $< -5.58\text{ dB}$ | **28%** |
| $< -11.46\text{ dB}$ | **7%** | $< -5.09\text{ dB}$ | **30%** |
| $< -10.68\text{ dB}$ | **8%** | $< -4.47\text{ dB}$ | **35%** |
| $< -9.80\text{ dB}$ | **10%** | $< -3.92\text{ dB}$ | **40%** |
| $< -8.92\text{ dB}$ | **12%** | $< -3.57\text{ dB}$ | **45%** |
| $\ge -3.57\text{ dB}$ | **50%** *(1:2 Simetris)* | — | — |

---

## 5. Manajemen Pelanggan, Layanan & Sistem Isolir Otomatis

### 5.1 Layanan PPPoE
- **Halaman**: `/pppoe_mikrotik` & `/user_list`
- **Atribut Data PPPoE**:
  - `Username`: ID akun PPPoE unik pelanggan.
  - `Password`: Sandi otentikasi dial-up PPPoE.
  - `Service`: `pppoe` (atau `any`).
  - `Profile`: Nama PPP Profile MikroTik yang menentukan alokasi IP Pool, DNS Server, dan Queue Limit.
  - `Local Address`: Gateway IP interface pppoe server.
  - `Remote Address`: IP statis / dinamis yang diberikan ke ONT klien.
  - `ONU Index & Serial Number`: Relasi identitas fisik modem pada port PON OLT.
- **Operasi Khusus**:
  - `check_koneksi_pppoe`: Memeriksa status dial PPPoE (Active vs Disconnected).
  - `restart_koneksi_pppoe`: Mengirim perintah `/ppp active remove [find name="..."]` untuk memaksa modem melakukan re-dial instan.
  - `modal_linked_onu`: Menghubungkan akun PPPoE dengan port OLT sehingga status LOS atau redaman optic dapat dimunculkan langsung pada baris pelanggan.

---

### 5.2 Layanan Hotspot Member & Voucher
- **Halaman**: `/hotspot_mikrotik`
- **Atribut Data Hotspot**:
  - `Server`: Nama Hotspot Server MikroTik (misal: `all` atau `hs-server1`).
  - `Username` & `Password`: Kredensial login captive portal.
  - `Profile`: Hotspot User Profile (menentukan Shared Users, Rate Limit, Session Timeout, Keepalive Timeout).
  - `Mac Address`: MAC binding untuk login otomatis (*Bypass/Cookie*).
  - `Data Limit`: Kuota kuantitas trafik (misal: 10 GB).
  - `Uptime Limit`: Batas durasi pemakaian (misal: 24 jam).
- **Fitur Mass Expired**:
  Sistem mengeksekusi penghapusan massal voucher kedaluwarsa via script MikroTik:
  `/ip hotspot user remove [find comment="..."]`.

---

### 5.3 Layanan IP Static
- **Halaman**: `/list_ip_static` & `/add_ip_static`
- **Atribut & Parameter Simple Queue Lengkap**:
  Setiap pelanggan IP Static langsung dibuatkan entri Simple Queue di MikroTik dengan parameter terperinci:
  1. `Target IP`: Alamat IP pelanggan (misal: `192.168.88.250/32`).
  2. `Max Limit (Upload/Download)`: Kecepatan puncak reguler (misal: `5M/10M`).
  3. `Limit At (Upload/Download)`: Kecepatan minimum bergaransi (CIR - Committed Information Rate).
  4. `Burst Limit (Upload/Download)`: Kecepatan lonjakan maksimal di awal pemakaian data (misal: `15M/30M`).
  5. `Burst Threshold (Upload/Download)`: Ambang batas trafik rata-rata untuk memicu burst.
  6. `Burst Time (Upload/Download)`: Periode perhitungan rata-rata burst (dalam detik).
  7. `Priority`: Skala prioritas paket `1` (tertinggi) hingga `8` (terendah).
  8. `Bucket Size`: Besaran buffer penampung burst (default: `0.1`).
  9. `Queue Type`: Jenis antrian (`pcq-upload-default`, `pcq-download-default`, `fq_codel`, `cake`).
  10. `Parent Queue`: Induk antrian bandwidth untuk manajemen bandwidth bertingkat (*hierarchical queuing*).

---

### 5.4 Deep Dive Sistem Isolir Otomatis

Sistem isolir Global SmartApp adalah salah satu mekanisme paling penting dalam arsitektur billing.

#### A. Alokasi Jaringan Isolir MikroTik:
- **Subnet Isolir**: `17.17.21.0/24`
- **Gateway / IP WebProxy**: `17.17.21.1`
- **Port WebProxy Isolir**: `3125`
- **Address-List Pelanggan Menunggak**: `ISOLIR`
- **Address-List Whitelist Server Billing**: `IPBILL-GLOBAL`

#### B. Script Konfigurasi Firewall & WebProxy MikroTik (Reverse Engineered):

```routeros
# ============================================================
# 1. SETUP WEB PROXY INTERNAL MIKROTIK UNTUK ISOLIR
# ============================================================
# Tentukan path penyimpanan cache error.html
:global wppath "webproxy";
:if ([:len [/file find name="flash"]] > 0) do={ :set wppath "flash/webproxy" };

/ip proxy access remove [find src-address~"17.17.21"];
/file remove [find name=($wppath . "/error.html")];

# Aktifkan Web Proxy pada port 3125
/ip proxy set cache-path=$wppath enabled=yes port=3125 max-cache-size=none;

# Aturan Akses Web Proxy:
# Redirect semua request browsing pelanggan selain ke IP 17.17.21.1:3125
/ip proxy access add action=deny dst-host=!17.17.21.1 redirect-to=17.17.21.1:3125 src-address=17.17.21.0/24;
/ip proxy access add dst-port=3125 src-address=17.17.21.0/24;
/ip proxy access add action=deny dst-address=!17.17.21.1 redirect-to=17.17.21.1:3125 src-address=17.17.21.0/24;

# ============================================================
# 2. FIREWALL ADDRESS-LIST & WHITELIST BILLING
# ============================================================
/ip firewall address-list remove [find comment="GLOBAL-ISOLIR"];
/ip firewall address-list add address=wg01.global-smartapp.com list=IPBILL-GLOBAL comment="GLOBAL-ISOLIR";
/ip firewall address-list add address=tun01.global-smartapp.com list=IPBILL-GLOBAL comment="GLOBAL-ISOLIR";
/ip firewall address-list add address=isolir.global-smartapp.com list=IPBILL-GLOBAL comment="GLOBAL-ISOLIR";

# ============================================================
# 3. FIREWALL NAT REDIRECT ISOLIR
# ============================================================
/ip firewall nat remove [find comment="GLOBAL-ISOLIR"];
# Redirect request HTTP (port 80) & HTTPS (port 443) ke port Proxy 3125
/ip firewall nat add action=redirect chain=dstnat comment="GLOBAL-ISOLIR" dst-port=80,443 protocol=tcp src-address-list=ISOLIR to-ports=3125;
/ip firewall nat add action=dst-nat chain=dstnat comment="GLOBAL-ISOLIR" dst-port=80,443 protocol=tcp src-address-list=ISOLIR to-addresses=17.17.21.1 to-ports=3125;

# ============================================================
# 4. FIREWALL FILTER FORWARD ISOLIR
# ============================================================
/ip firewall filter remove [find comment="GLOBAL-ISOLIR"];
# A. Izinkan koneksi established & related
/ip firewall filter add action=accept chain=forward comment="GLOBAL-ISOLIR" connection-state=established,related place-before=0;

# B. Izinkan DNS Port 53 (UDP & TCP) agar resolusi domain tetap bisa berjalan
/ip firewall filter add action=accept chain=forward comment="GLOBAL-ISOLIR" dst-port=53 protocol=udp src-address-list=ISOLIR place-before=0;
/ip firewall filter add action=accept chain=forward comment="GLOBAL-ISOLIR" dst-port=53 protocol=tcp src-address-list=ISOLIR place-before=0;

# C. Izinkan akses ke server billing & payment gateway (IPBILL-GLOBAL)
/ip firewall filter add action=accept chain=forward comment="GLOBAL-ISOLIR" dst-address-list=IPBILL-GLOBAL src-address-list=ISOLIR place-before=0;

# D. BLOKIR / DROP semua trafik internet lainnya untuk list ISOLIR
/ip firewall filter add action=drop chain=forward comment="GLOBAL-ISOLIR" src-address-list=ISOLIR place-before=0;
```

#### C. Dua Mode Eksekusi Halaman Isolir:
1. **Mode Self (Upload ke MikroTik)**:
   File template `error.html` di-upload via FTP lokal ke folder flash MikroTik. Ketika pelanggan membuka browser, MikroTik sendiri yang menyajikan halaman isolir secara lokal tanpa memakan bandwidth internet.
2. **Mode Server Isolir Global**:
   Trafik diarahkan ke server cloud `https://isolir.global-smartapp.com`. Mode ini memudahkan pembaruan tampilan isolir secara serentak dari cloud tanpa perlu mengunggah ulang file ke setiap router.

---

### 5.5 Workflow Planning PSB (Pasang Baru Mandiri)

Modul `/planning_psb` mengelola siklus hidup pemasangan pelanggan baru dari registrasi hingga aktivasi:

```
[ Link Publik / QRIS Mandiri ]
              │
              ▼
   [ 1. Status: PENDING ]
   - Form pendaftaran online diisi calon pelanggan / sales lapangan.
   - Mengumpulkan data: Nama, No WhatsApp, NIK, Foto Rumah, Share Location GPS.
              │
              ▼
    [ 2. Status: SURVEY ]
   - Teknisi memeriksa jarak ke tiang ODP terdekat melalui modul GIS.
   - Pengecekan ketersediaan port kosong pada ODP.
              │
              ▼
  [ 3. Status: ON PROGRESS ]
   - Penarikan kabel dropcore dari ODP ke rumah pelanggan.
   - Splicing roset & penyambungan konektor optik ke ONT.
              │
              ▼
   [ 4. Status: SELESAI / AKTIF ]
   - Dialog Modal `modalSelesaikanPSB` dipicu.
   - Sistem otomatis membuatkan akun PPPoE / IP Static di MikroTik.
   - Registrasi MAC / SN ONT ke OLT via Telnet / ACS.
   - Invoice instalasi PSB otomatis terkirim ke WhatsApp pelanggan.
```

---

### 5.6 Prorata Billing & Diskon

1. **Prorata (`/prorata`)**:
   Jika pelanggan mulai aktif pada tanggal 20 dengan tanggal jatuh tempo bulanan tanggal 1:
   $$\text{Tagihan Prorata} = \frac{\text{Sisa Hari Aktif}}{\text{Total Hari Bulan Berjalan}} \times \text{Harga Paket Bulanan}$$
2. **Diskon (`/discount`)**:
   Mendukung diskon persentase (%) atau potongan nominal tetap (Rp) yang dapat diterapkan pada biaya pasang baru (PSB) atau potongan tagihan berkala.

---

## 6. Sistem Billing, Keuangan & Payment Gateway

### 6.1 Payment Gateway Global Moneta
Global Moneta adalah channel payment gateway terintegrasi bawaan Global SmartApp:

#### A. Parameter Rekening Pencairan (Withdraw):
- `jenis_bank`: Pilihan bank tujuan pencairan dana (BCA, BNI, BRI, Mandiri, Permata, BSS, BNC, CIMB Niaga, Danamon, BSI, BJB, Panin, Mega, OCBC NISP, Sinarmas, Maybank, UOB).
- `nomor_rekening`: Nomor rekening bank milik mitra.
- `nama_rekening`: Nama pemilik rekening.
- **Validasi Rekening Otomatis (`init_account_validation_moneta`)**: Sistem melakukan pengecekan nama pemilik rekening secara real-time via API Moneta begitu nomor rekening selesai diketik untuk menghindari salah transfer.

#### B. Skema Pembebanan Biaya Admin:
1. **100% Ditanggung Mitra**: Biaya penanganan gateway langsung memotong saldo penghasilan mitra.
2. **100% Ditanggung Pelanggan**: Tagihan pelanggan ditambahkan biaya admin gateway (misal Rp 100.000 + Rp 2.500 = Rp 102.500).
3. **50% Mitra & 50% Pelanggan**: Biaya admin dibagi dua secara berimbang.
4. **Biaya Admin Tambahan (Markup Komisi Mitra)**: Fitur menambahkan biaya layanan kustom di luar biaya gateway sebagai pendapatan ekstra bagi ISP.

#### C. Instant Un-isolir (Webhook Callback):
Ketika pelanggan membayar tagihan melalui QRIS atau Virtual Account:
1. Server Global Moneta mengirimkan HTTP POST Webhook ke sistem billing.
2. Sistem billing mencatat transaksi lunas di database.
3. Server billing mengeksekusi API MikroTik:
   `/ip firewall address-list remove [find address="IP_PELANGGAN" list="ISOLIR"]`
4. Internet pelanggan langsung kembali normal dalam hitungan detik tanpa campur tangan admin (*zero-touch automation*).

---

### 6.2 Integrasi Midtrans
Untuk mitra yang telah memiliki akun payment gateway mandiri:
- **Mode Sistem**: Pilihan `Sandbox` (Pengujian) atau `Production` (Live).
- **Kredensial**:
  - `ID Merchant`
  - `Client Key`
  - `Server Key`
- Mendukung metode pembayaran GoPay, ShopeePay, QRIS, dan Virtual Account Bank Transfer dengan modul popup Midtrans Snap.

---

### 6.3 Sistem Pembayaran Manual, Upload Bukti Transfer & Approval
- Pelanggan melakukan transfer konvensional ke rekening bank mitra.
- Pelanggan/kolektor mengunggah foto struk/screenshot mutasi transfer.
- Pada halaman `/user_list` atau `/list_payment`, admin mengklik dialog modal **`#modaldetailpppoe`**:
  - Menampilkan: Order ID, Tanggal Bayar, Nama Klien, Bank Tujuan, Nominal, PPN, dan foto Bukti Transfer yang dapat di-zoom (`modal-zoom-image`).
  - Tombol **Setujui**: Menandai transaksi lunas, memperpanjang tanggal isolir ke bulan berikutnya, dan menghapus IP pelanggan dari address-list `ISOLIR` MikroTik.
  - Tombol **Batalkan**: Menolak pembayaran jika transfer tidak valid.

---

### 6.4 Struktur Pembukuan Arus Kas (Finance Ledger)
- **Income (`/income/home`)**: Arus kas masuk yang dikategorikan berdasarkan sumber: Pembayaran Bulanan PPPoE, Hotspot, IP Static, Biaya PSB, Penjualan Voucher Hotspot, dan Pendapatan Non-langganan.
- **Pengeluaran (`/pengeluaran/home`)**: Arus kas keluar untuk operasional ISP: Pembelian Bandwidth Uplink (Upstream/IX), Sewa Tiang/Kabel, Gaji Staf, Pemeliharaan Perangkat.
- **Hutang-Piutang (`/hutang_piutang`)**: Pelacakan saldo tertunggak dari pelanggan (*Accounts Receivable*) dan kewajiban pembayaran tempo ke vendor perangkat/supplier (*Accounts Payable*).

---

## Rangkuman Spesifikasi Teknis 6 Modul

| No | Modul | Endpoint Utama | Protokol / Port Kunci | Mekanisme Otomatisasi Utama |
|---|---|---|---|---|
| **1** | **Arsitektur Jaringan** | `/admin`, `/account` | WireGuard (UDP 13230), OpenVPN (TCP 1194) | Cloud tunnel hub bypass NAT tanpa IP Publik statis |
| **2** | **Remote Device** | `/remote_device` | HTTP Reverse Proxy Tunnel | Proxy Web GUI ONT pelanggan via port dinamis |
| **3** | **Network Core** | `/routeros/*`, `/olt/*`, `/acs` | API (8728), Telnet (23), SNMP (161), CWMP (321) | Watchdog scheduler, OLT auto-save, telemetri RX optik |
| **4** | **Topologi FTTH** | `/calc_ratio/*`, `/topologi` | GIS Leaflet, Splitting Math JS | Kalkulator redaman dBm optik (kabel 0.35dB/km, konektor, splicing) |
| **5** | **Pelanggan & Isolir** | `/user_list`, `/home_isolir` | WebProxy (3125), NAT DST-NAT, Address-List | Auto-redirect penunggak ke halaman isolir, PPPoE re-dial |
| **6** | **Billing & Payment** | `/payment_gateway`, `/list_payment` | Webhook HTTP POST, Rest API Moneta/Midtrans | Instant un-isolir via webhook callback QRIS & VA realtime |
