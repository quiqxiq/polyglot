# Panduan Deployment GNet Dial dengan Nginx

Panduan ini menjelaskan cara melakukan deployment **GNet Dial** pada server produksi menggunakan subdomain terpisah (contoh: `dial.ispanda.com`) dengan reverse proxy Nginx yang terhubung ke backend Go **Polyglot**.

---

## 1. Topologi Layanan

```text
               +----------------------------------------+
               |             Client Browser             |
               +----------------------------------------+
                       |                        |
             dial.ispanda.com            app.ispanda.com
             (Portal Teknisi)             (Web Admin ISP)
                       |                        |
                       v                        v
             +------------------+     +------------------+
             |   Nginx Server   |     |   Nginx Server   |
             |   (dial.conf)    |     |   (app.conf)     |
             +------------------+     +------------------+
               |              \                 |
         Static Files        /polyglot.v1.*     |
      /var/www/.../web_dial          \          |
                                      v         v
                             +--------------------+
                             | Polyglot Go Server |
                             |    (:8080)         |
                             +--------------------+
                                      |
                             +--------------------+
                             | PostgreSQL / Redis |
                             +--------------------+
```

---

## 2. Langkah Konfigurasi Nginx

### Langkah 1: Pasang Konfigurasi Virtual Host

Salin file `nginx-dial.conf` dari repository ke direktori konfigurasi Nginx:

```bash
sudo cp /var/www/polyglot/web_dial/nginx-dial.conf /etc/nginx/sites-available/dial.ispanda.com.conf
```

### Langkah 2: Sesuaikan Domain & Path

Buka `/etc/nginx/sites-available/dial.ispanda.com.conf` dan pastikan:
1. `server_name` sesuai dengan subdomain yang Anda inginkan (misal `dial.ispanda.com`).
2. `root` mengarah ke lokasi folder `web_dial` di server Anda (misal `/var/www/polyglot/web_dial`).
3. `proxy_pass` mengarah ke alamat backend Go Polyglot (biasanya `http://127.0.0.1:8080`).

### Langkah 3: Aktifkan Virtual Host

```bash
sudo ln -s /etc/nginx/sites-available/dial.ispanda.com.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Langkah 4: Pasang Sertifikat SSL (HTTPS) dengan Certbot

```bash
sudo certbot --nginx -d dial.ispanda.com
```

---

## 3. Menjalankan Backend Polyglot

Pastikan service `polyglot` aktif dan mendengarkan pada port 8080:

```bash
# Cek status service
sudo systemctl status polyglot

# Atau uji koneksi lokal
curl http://127.0.0.1:8080/polyglot.v1.DeviceService/ListDevices
```

---

## 4. Keuntungan Arsitektur Ini

* **Zero Maintenance PHP/MySQL**: Tidak perlu lagi menginstal `php-fpm`, ekstensi routeros php, atau menjalankan MySQL/MariaDB terpisah untuk portal monitoring.
* **Performa Tinggi**: Aset statis disajikan langsung oleh Nginx secara optimal dengan kompresi Gzip dan HTTP/2.
* **Keamanan Kredensial Terpusat**: Password router MikroTik tidak disimpan di file konfigurasi web publik, melainkan terenkripsi dengan AES-256 GCM di database PostgreSQL backend Polyglot.
