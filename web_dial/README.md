# GNet Dial - Portal Operasional Teknisi & Admin

Portal web statis (Single/Multi-Page Application) khusus untuk **Teknisi Lapangan** dan **Admin** dalam memantau dan mengoperasikan layanan PPP MikroTik. Terhubung langsung ke backend Go **Polyglot** melalui ConnectRPC (JSON over HTTP), tanpa dependensi PHP maupun database MySQL mandiri.

---

## Fitur Utama

1. **Dashboard Monitoring**: Menampilkan identitas router (Board Name, RouterOS Version, Uptime, CPU Load bar, Memori RAM terpakai), serta form pencarian cepat PPP.
2. **PPP Secrets**: Daftar seluruh pelanggan PPPoE / Hotspot yang terdaftar pada router yang dipilih, lengkap dengan filter nama, profile, dan status enabled/disabled.
3. **PPP Aktif**: Pemantauan pelanggan yang sedang online secara realtime, dilengkapi dengan:
   * Tombol Putus Sesi (Kick) single & batch.
   * Modal Uji Latensi Jaringan (Ping Test).
   * Tombol langsung buka antarmuka CPE / WebFig pelanggan di tab baru.
4. **PPP Non-Aktif**: Daftar akun pelanggan yang sedang offline beserta waktu logout terakhir dan komentarnya.
5. **PPP Profiles**: Daftar profil rate-limit dan bandwidth package pada router.
6. **Log MikroTik**: Streaming log sistem router MikroTik secara realtime dengan filter topik (pppoe, warning, system).
7. **Monitor Interface**: Pemantauan traffic rate (RX/TX) dan link status pada setiap port ethernet fisik serta queue.
8. **Pengaturan (Khusus Admin)**:
   * Halaman ini **hanya dapat diakses dan terlihat oleh akun ber-role `admin`**.
   * Admin dapat membuat akun `teknisi` baru.
   * Admin dapat menentukan router-router mana saja (dari daftar router yang dimiliki Admin) yang boleh diakses oleh teknisi tersebut.

---

## Kontrol Hak Akses (RBAC)

* **Superadmin / Owner**: Beroperasi di portal utama ISP (`web/`), tidak dapat login ke portal ini (diblokir oleh guard login dengan notifikasi terarah).
* **Admin**:
  * Mengakses router-router yang di-assign kepadanya.
  * Mengelola akun teknisi dan memilihkan router untuk mereka melalui menu **Pengaturan**.
* **Teknisi**:
  * Hanya melihat dan memilih router yang telah di-assign oleh Admin untuknya.
  * Menu **Pengaturan** disembunyikan dan diproteksi.

---

## Struktur Direktori

```text
web_dial/
├── index.html                  # Dashboard System Resource & Quick Search
├── ppp-secrets.html            # Daftar PPP Secrets
├── ppp-active.html             # PPP Aktif, Kick Session, Modal Ping, Link CPE
├── ppp-non-active.html         # PPP Offline
├── ppp-profiles.html           # PPP Profiles
├── mikrotik-logs.html          # Log Sistem MikroTik Realtime
├── monitor-interface.html      # Monitor Interface Ethernet
├── settings.html               # Pengaturan: Manajemen Teknisi & Router (Khusus Admin)
├── login.html                  # Autentikasi dengan Guard Role
├── assets/                     # Icons & Favicon
├── js/
│   ├── api.js                  # ConnectRPC JSON Client Wrapper
│   ├── auth.js                 # JWT Session & Role Guard
│   ├── layout.js               # Navigasi Sidebar & Header Dinamis
│   ├── router-selector.js      # Global Header Router Picker
│   ├── dashboard.js            # Controller Dashboard
│   ├── ppp-secrets.js          # Controller PPP Secrets
│   ├── ppp-active.js           # Controller PPP Aktif
│   ├── ppp-non-active.js       # Controller PPP Offline
│   ├── ppp-profiles.js         # Controller PPP Profiles
│   ├── mikrotik-logs.js        # Controller Log Realtime
│   ├── monitor-interface.js    # Controller Monitor Interface
│   └── settings.js             # Controller Pengaturan Teknisi & Router
└── nginx-dial.conf             # Konfigurasi Nginx Reverse Proxy
```

---

## Cara Menjalankan

### 1. Di Lingkungan Development (Lokal)
Karena portal ini murni file statis (HTML/CSS/JS):
* Jalankan backend Polyglot: `go run ./cmd/server` (berjalan di port `8080`).
* Jalankan server statis sederhana pada folder `web_dial/`, misalnya dengan Nginx, Caddy, atau Live Server:
  ```bash
  # Contoh menggunakan Python HTTP server
  cd web_dial
  python3 -m http.server 3000
  ```
* Buka browser di `http://localhost:3000/login.html`.

### 2. Di Lingkungan Production (Nginx Reverse Proxy)
* Salin file `nginx-dial.conf` ke `/etc/nginx/sites-available/dial.conf`.
* Sesuaikan `server_name` dengan domain Anda (misal `dial.ispanda.com`) dan arahkan `root` ke direktori `web_dial`.
* Aktifkan site dan reload Nginx:
  ```bash
  sudo ln -s /etc/nginx/sites-available/dial.conf /etc/nginx/sites-enabled/
  sudo nginx -t && sudo systemctl reload nginx
  ```
