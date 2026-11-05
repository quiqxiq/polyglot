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
8. **Pengaturan (berbasis permission)**:
   * Menu dan tab hanya tampil bila akun memiliki permission yang sesuai.
   * Tab **Akun Teknisi** membutuhkan `user:read`; pembuatan, perubahan status, dan penugasan router membutuhkan `user:manage`.
   * Tab **Akses Router** membutuhkan `user:manage` **dan** `device:read`.

---

## Kontrol Hak Akses (RBAC)

Portal ini memakai RBAC berbasis menu yang diturunkan langsung dari permission efektif akun (daftar yang sama dengan yang ditegakkan Casbin di backend). Tidak ada allowlist role di frontend.

* **Owner / Superadmin**: wildcard `.*` — seluruh menu, halaman, dan aksi Dial tersedia.
* **Admin**: seluruh resource kecuali `rbac:manage`; dapat mengelola akun teknisi dan akses router.
* **Teknisi**: `device:read`, `device:diagnostic:exec`, `ppp:secret:read`, `ppp:profile:read`, `ppp:active:read`, `monitor:read`, dan `log:read`. Menu Pengaturan, tombol Kick, dan toggle secret tidak tersedia.
* **Agent**: default tidak memiliki menu Dial karena tidak memiliki `device:read`/`user:read`. Akses dapat diberikan kapan saja dengan menambah permission; kode menu tidak perlu diubah.

### Prinsip cascade

* Menolak sebuah node menu otomatis menolak seluruh descendant-nya, walaupun permission anak terpenuhi.
* Setiap halaman memanggil `Auth.requireMenu(...)` saat bootstrap; halaman yang tidak diizinkan diarahkan ke `no-access.html` atau rute pertama yang boleh diakses.
* Tombol aksi memakai `Auth.canAction(...)`; API tetap menjadi batas keamanan, sehingga respons 403 menampilkan notifikasi sekali per prosedur dan memicu event `api:permission-denied`.
* Pencocokan wildcard di client di-anchor (`^(?:pattern)$`) agar policy seperti `profile:.*` tidak bocor ke `ppp:profile:read`, konsisten dengan `regexMatch(r.obj, "^" + p.obj + "$")` di Casbin.

### Hierarki menu

```text
Dashboard (device:read)
Operasional (device:read)
├── PPP (salah satu: ppp:active:read / ppp:secret:read / ppp:profile:read)
│   ├── PPP Aktif (ppp:active:read)
│   ├── PPP Non-Aktif (ppp:active:read)
│   ├── PPP Secrets (ppp:secret:read)
│   └── PPP Profiles (ppp:profile:read)
└── Monitoring (salah satu: log:read / monitor:read)
    ├── Log MikroTik (log:read)
    └── Monitor Interface (monitor:read + device:diagnostic:exec)
Sistem
└── Pengaturan (user:read)
    ├── Akun Teknisi (user:read)
    └── Akses Router (user:manage + device:read)
```

Role `technician` legacy dimigrasikan ke `teknisi` melalui migration `000041_technician_role_merge`. Login dan refresh token merekonsiliasi assignment Casbin dengan kolom `users.role`; role kustom yang dibuat lewat layanan RBAC tetap dipertahankan.

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
├── settings.html               # Pengaturan: tab Akun Teknisi & Akses Router
├── login.html                  # Autentikasi dengan admission berbasis menu
├── no-access.html              # Halaman penolakan akses
├── assets/                     # Icons & Favicon
├── js/
│   ├── api.js                  # ConnectRPC JSON Client Wrapper + penanganan 403
│   ├── auth.js                 # JWT Session, permission, menu & action guard
│   ├── menu.js                 # Registry menu + resolver cascade (sumber kebenaran)
│   ├── menu.test.js            # Unit test resolver (node --test)
│   ├── layout.js               # Navigasi sidebar/drawer/bottom dari menu efektif
│   ├── router-selector.js      # Global Header Router Picker (butuh device:read)
│   ├── i18n.js / locales.js    # Internasionalisasi ID/EN
│   ├── dashboard.js            # Controller Dashboard
│   ├── ppp-secrets.js          # Controller PPP Secrets
│   ├── ppp-active.js           # Controller PPP Aktif
│   ├── ppp-non-active.js       # Controller PPP Offline
│   ├── ppp-profiles.js         # Controller PPP Profiles
│   ├── mikrotik-logs.js        # Controller Log Realtime
│   ├── monitor-interface.js    # Controller Monitor Interface
│   └── settings.js             # Controller tab Pengaturan
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

### 3. Menjalankan Unit Test Menu
Resolver menu bersifat murni dan dapat diuji tanpa browser:

```bash
node --test web_dial/js/menu.test.js
```

