## 1. PPP Secret (User PPPoE)

### `/ppp/secret/print`
Filter dipakai di kode: `?.name=<user>`, `?name=<user>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID internal, dipakai untuk set/remove |
| `name` | Username PPPoE |
| `password` | Password |
| `profile` | Nama profile PPP |
| `service` | Jenis service (pppoe/any/dll) |
| `caller-id` | MAC/Caller ID terakhir |
| `local-address` | IP lokal statis (jika diset) |
| `remote-address` | IP remote statis / nama pool |
| `limit-bytes-in` / `limit-bytes-out` | Kuota byte |
| `last-logged-out` | Waktu logout terakhir |
| `last-caller-id` | Caller ID sesi terakhir |
| `last-disconnect-reason` | Alasan disconnect terakhir |
| `comment` | Komentar |
| `disabled` | Status aktif/nonaktif |

**Field yang dibaca aplikasi:** `.id`, `name`, `password`, `profile`, `comment`, `last-logged-out`

### `/ppp/secret/add`
```
=name=<username>
=password=<password>
=profile=<profile>
=service=pppoe
=local-address=<local_address>   // opsional, hanya jika diisi
```

### `/ppp/secret/set`
Dipakai di 2 skenario berbeda:
```
// editPPPoEUser — ubah username/password/profile
=.id=<id>
=name=<username>
=password=<password>
=profile=<profile>

// setPPPoEProfile — ganti profile saja
=.id=<id>
=profile=<profile>

// serviceSuspension (isolir) — suspend pelanggan
=.id=<id>              // fallback: =name=<username> jika id tidak ditemukan
=profile=<isolir_profile>
=comment=SUSPENDED - <reason>
```

### `/ppp/secret/remove`
```
=.id=<id>
```

---

## 2. PPP Active (Sesi PPPoE Aktif) — Monitoring

### `/ppp/active/print`
Filter dipakai: `?name=<user>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID sesi aktif |
| `name` | Username |
| `service` | Jenis service |
| `caller-id` | MAC address client |
| `address` | IP yang diberikan ke client |
| `uptime` | Durasi koneksi |
| `encoding` | Info enkripsi/kompresi |
| `session-id` | ID sesi |
| `limit-bytes-in` / `limit-bytes-out` | Limit byte sesi |
| `radius` | Apakah via RADIUS |

**Field yang dibaca aplikasi:** `.id`, `name` (untuk pencocokan status online/offline & kick)

### `/ppp/active/remove` (kick/disconnect user)
```
=.id=<id>
```
Dipanggil setelah lookup by `?name=<username>` — dipakai saat ganti profile, saat isolir, dan pada `kickPPPoEUser` (`mikrotik2.js`).

---

## 3. PPP Profile

### `/ppp/profile/print`
Filter dipakai: `?.id=<id>`, `?name=<name>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID profile |
| `name` | Nama profile |
| `local-address` | IP gateway PPP |
| `remote-address` | IP/pool untuk client |
| `rate-limit` | Limit bandwidth (rx/tx) |
| `parent-queue` | Parent queue untuk shaping |
| `dns-server` | DNS yang didorong ke client |
| `address-list` | Address-list otomatis untuk IP client |
| `bridge-learning` | Mode bridge |
| `use-mpls` / `use-compression` / `use-encryption` | Flag koneksi |
| `only-one` | Batasi 1 sesi per user |
| `change-tcp-mss` | Penyesuaian MSS |
| `comment` | Komentar |

### `/ppp/profile/add`
```
=name=<name>
=rate-limit=<value>          // opsional
=local-address=<value>       // opsional
=remote-address=<value>      // opsional
=dns-server=<value>          // opsional
=parent-queue=<value>        // opsional
=address-list=<value>        // opsional
=comment=<value>             // opsional
=bridge-learning=<value>     // opsional, hanya jika != 'default'
=use-mpls=<value>            // opsional, hanya jika != 'default'
=use-compression=<value>     // opsional, hanya jika != 'default'
=use-encryption=<value>      // opsional, hanya jika != 'default'
=only-one=<value>            // opsional, hanya jika != 'default'
=change-tcp-mss=<value>      // opsional, hanya jika != 'default'
```

Kasus khusus — auto-create profile isolir (`serviceSuspension.js → ensureIsolirProfile`):
```
=name=isolir
=local-address=0.0.0.0
=remote-address=0.0.0.0
=rate-limit=0/0
=comment=SUSPENDED_PROFILE
=shared-users=1
```
> Response `add` dibaca lewat `result[0]['ret']` untuk mendapatkan `.id` profile baru.

### `/ppp/profile/set`
```
=.id=<id>
=name=<name>              // jika diisi
=rate-limit=<value>       // hanya field yang dikirim user, pola: if (value !== undefined) push
=local-address=<value>
=remote-address=<value>
=dns-server=<value>
=parent-queue=<value>
=address-list=<value>
=comment=<value>
=bridge-learning=<value>
=use-mpls=<value>
=use-compression=<value>
=use-encryption=<value>
=only-one=<value>
=change-tcp-mss=<value>
```

### `/ppp/profile/remove`
```
=.id=<id>
```

---

## 4. Hotspot User

### `/ip/hotspot/user/print`
Filter dipakai: `?name=<username>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID user |
| `server` | Server hotspot terkait |
| `name` | Username |
| `password` | Password |
| `profile` | Profile rate-limit |
| `mac-address` | MAC binding |
| `address` | IP statis (jika ada) |
| `limit-uptime` | Batas waktu total |
| `limit-bytes-in` / `limit-bytes-out` | Kuota byte |
| `comment` | Komentar (dipakai app untuk tag `voucher`) |
| `disabled` | Status |

### `/ip/hotspot/user/add`
```
=name=<username>
=password=<password>
=profile=<profile>
=comment=<comment>     // opsional
=server=<server>       // opsional, hanya jika bukan 'all' (voucher generator)
```

### `/ip/hotspot/user/set`
```
=numbers=<id>          // ⚠️ pakai key "numbers", BUKAN ".id"
=password=<password>
=profile=<profile>
```
> Catatan: fungsi `updateHotspotUser` di `mikrotik.js` menggunakan `=numbers=` (bukan `=.id=`) untuk menunjuk item — berbeda dari pola set di endpoint lain dalam codebase yang sama.

### `/ip/hotspot/user/remove`
```
=.id=<id>
```

---

## 5. Hotspot Active (Sesi Aktif) — Monitoring

### `/ip/hotspot/active/print`
Filter dipakai: `?user=<username>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID sesi |
| `server` | Server hotspot |
| `user` | Username |
| `address` | IP client |
| `mac-address` | MAC client |
| `login-by` | Metode login (http-pap/cookie/dll) |
| `uptime` | Durasi sesi |
| `session-time-left` | Sisa waktu sesi |
| `idle-time` | Waktu idle |
| `bytes-in` / `bytes-out` | Trafik terpakai |
| `packets-in` / `packets-out` | Jumlah paket |

**Field yang dibaca aplikasi:** `.id`, seluruh array dipakai langsung sebagai data user aktif (tanpa mapping tambahan)

### `/ip/hotspot/active/remove` (disconnect user)
```
=.id=<id>
```

---

## 6. Hotspot User Profile

### `/ip/hotspot/user/profile/print`
Filter dipakai: `?.id=<id>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID profile |
| `name` | Nama profile |
| `rate-limit` | Limit bandwidth |
| `session-timeout` | Batas durasi sesi |
| `idle-timeout` | Batas waktu idle |
| `shared-users` | Jumlah device bersamaan |
| `parent-queue` | Parent queue |
| `comment` | Komentar |

### `/ip/hotspot/user/profile/add`
```
=name=<name>
=comment=<comment>              // opsional
=rate-limit=<rateLimit+unit>    // opsional
=session-timeout=<value+unit>   // opsional
=idle-timeout=<value+unit>      // opsional
=local-address=<value>          // opsional
=remote-address=<value>         // opsional
=dns-server=<value>             // opsional
=parent-queue=<value>           // opsional
=address-list=<value>           // opsional
=shared-users=<value>           // opsional
```
> ⚠️ **Catatan analisis:** `local-address`, `remote-address`, `dns-server`, dan `address-list` **bukan properti resmi** `/ip/hotspot/user/profile` di RouterOS (itu properti milik `/ppp/profile`). Properti resmi untuk pool di menu ini adalah `address-pool`. Field-field tersebut kemungkinan besar akan diabaikan/ditolak oleh router tergantung versi RouterOS — perlu diverifikasi langsung di router target.

### `/ip/hotspot/user/profile/set`
```
=.id=<id>
=name=<name>
=comment=<comment>              // jika !== undefined
=rate-limit=<value>             // atau '=rate-limit=' (kosong) untuk clear
=session-timeout=<value>        // atau kosong untuk clear
=idle-timeout=<value>           // atau kosong untuk clear
=local-address=<value>
=remote-address=<value>
=dns-server=<value>
=parent-queue=<value>
=address-list=<value>
=shared-users=<value>
```

### `/ip/hotspot/user/profile/remove`
```
=.id=<id>
```

---

## 7. Hotspot Server (read-only di aplikasi ini)

### `/ip/hotspot/print`

| Field raw | Keterangan |
|---|---|
| `.id` | ID server |
| `name` | Nama server hotspot |
| `interface` | Interface terkait |
| `profile` | Server profile |
| `address-pool` | Pool IP untuk client (dipetakan aplikasi sebagai `address`) |
| `disabled` | Status |

> Tidak ada operasi `add`/`set`/`remove` untuk `/ip/hotspot` (server) di codebase ini — hanya dibaca (`getHotspotServers`).

---

## 8. VPN (PPTP/L2TP/SSTP/OVPN/WireGuard)

**Tidak diimplementasikan.** Tidak ada satupun pemanggilan ke `/interface/pptp-server`, `/interface/l2tp-server`, `/interface/sstp-server`, `/interface/ovpn-server`, atau `/interface/wireguard` di seluruh repo. Menu `/ppp/*` yang dipakai murni untuk **PPPoE** (secret, profile, active) — bukan modul VPN server terpisah.

---

## 9. IP Address

### `/ip/address/print`
Filter dipakai: `?interface=<name>`, `?address=<addr>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID entry |
| `address` | IP/CIDR |
| `network` | Network address |
| `interface` | Interface terkait |
| `disabled` | Status |
| `dynamic` | Apakah dari DHCP client |
| `comment` | Komentar |

### `/ip/address/add`
```
=interface=<interfaceName>
=address=<address>
```

### `/ip/address/remove`
```
=.id=<id>
```

---

## 10. IP Pool

**Tidak diimplementasikan sebagai modul CRUD terpisah.** Tidak ada pemanggilan ke `/ip/pool/print`, `/add`, `/set`, atau `/remove` di codebase. Pool hanya muncul secara tidak langsung:
- Sebagai field `remote-address` pada `/ppp/profile` (nama pool yang sudah dibuat manual di router).
- Sebagai field `address-pool` pada `/ip/hotspot/print` (dibaca, tidak pernah ditulis).

Manajemen pool (create/resize) diasumsikan dilakukan manual di router, di luar aplikasi.

---

## 11. IP Route

### `/ip/route/print`
Filter dipakai: `?dst-address=<dest>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID route |
| `dst-address` | Network tujuan |
| `gateway` | Gateway |
| `distance` | Administrative distance |
| `active` | Status aktif |
| `dynamic` / `static` | Sumber route |
| `comment` | Komentar |

### `/ip/route/add`
```
=dst-address=<destination>
=gateway=<gateway>
=distance=<distance>     // default '1'
```

### `/ip/route/remove`
```
=.id=<id>
```

---

## 12. DHCP Server & Lease

### `/ip/dhcp-server/print`

| Field raw | Keterangan |
|---|---|
| `.id` | ID server |
| `name` | Nama DHCP server |
| `interface` | Interface |
| `address-pool` | Pool IP |
| `lease-time` | Durasi lease |
| `disabled` | Status |

### `/ip/dhcp-server/lease/print`
Filter dipakai: `?mac-address=<mac>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID lease |
| `address` | IP yang di-lease |
| `mac-address` | MAC client |
| `client-id` | Client ID |
| `server` | DHCP server terkait |
| `status` | bound/waiting/dll |
| `active-address` | IP aktif |
| `host-name` | Hostname client |
| `blocked` | Status blokir |
| `comment` | Komentar |

### `/ip/dhcp-server/lease/set` (dipakai untuk suspend berbasis MAC)
```
=.id=<id>
=blocked=yes
=comment=SUSPENDED - <reason> - <timestamp>
```
Restore: sama, dengan `=blocked=no`.

> Tidak ada `add`/`remove` untuk DHCP server maupun lease — hanya `print` (monitoring) dan `set` (blokir/buka blokir lease existing).

---

## 13. Firewall Filter

### `/ip/firewall/filter/print`
Filter dipakai: `?chain=<chain>`, `?src-address=<ip>`, `?action=drop`, `?src-address-list=<list>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID rule |
| `chain` | forward/input/output |
| `action` | drop/accept/dll |
| `src-address` | Source IP |
| `src-address-list` | Referensi address-list |
| `protocol` | Protokol |
| `bytes` / `packets` | Counter trafik yang match |
| `disabled` | Status |
| `comment` | Komentar |

### `/ip/firewall/filter/add`
Tiga varian dipakai di kode (`staticIPSuspension.js`):
```
// 1. Block IP spesifik pelanggan
=chain=forward
=src-address=<customerIP>
=action=drop
=comment=SUSPENDED block_<ip> - <reason> - <timestamp>

// 2. Rule dasar block address-list (chain forward, dibuat sekali)
=chain=forward
=src-address-list=blocked_customers
=action=drop
=comment=Block suspended customers (static IP)
=place-before=0

// 3. Rule tambahan block address-list dari akses ke router (chain input)
=chain=input
=src-address-list=blocked_customers
=action=drop
=comment=Block suspended customers from accessing router (static IP)
```

### `/ip/firewall/filter/remove`
```
=.id=<id>
```

---

## 14. Firewall Address List

### `/ip/firewall/address-list/print`
Filter dipakai: `?list=<listName>`, `?address=<ip>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID entry |
| `list` | Nama address-list |
| `address` | IP |
| `comment` | Komentar |
| `disabled` | Status |
| `dynamic` | Sumber entry |
| `creation-time` | Waktu dibuat |

### `/ip/firewall/address-list/add`
```
=list=blocked_customers
=address=<customerIP>
=comment=SUSPENDED - <reason> - <timestamp>
```

### `/ip/firewall/address-list/remove`
```
=.id=<id>
```

---

## 15. Queue (Simple Queue) — Monitoring & Bandwidth Limit

### `/queue/simple/print`
Filter dipakai: `?name=<queueName>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID queue |
| `name` | Nama queue |
| `target` | IP/subnet target |
| `parent` | Parent queue (hierarki) |
| `max-limit` | Limit maksimum rx/tx |
| `limit-at` | CIR (guaranteed rate) |
| `burst-limit` / `burst-threshold` / `burst-time` | Parameter burst |
| `priority` | Prioritas (1–8) |
| `queue` | Tipe antrian (pfifo/fq_codel/dll) |
| `bytes` | Total byte terproses (rx/tx) |
| `packets` | Total paket terproses |
| `dropped` | Paket yang di-drop |
| `rate` | Rate saat ini (bps) |
| `packet-rate` | Rate saat ini (pps) |
| `disabled` | Status |
| `comment` | Komentar |

### `/queue/simple/add` (dipakai untuk soft-suspend/limit bandwidth pelanggan)
```
=name=suspended_<ip_underscored>
=target=<customerIP>
=max-limit=<limitSpeed>     // contoh default: 1k/1k
=comment=SUSPENDED - <reason> - <timestamp>
=disabled=no
```

### `/queue/simple/remove`
```
=.id=<id>
```

> Tidak ada penggunaan `/queue/tree` di codebase ini — seluruh queue management memakai **Simple Queue**.

---

## 16. Interface

### `/interface/print`
Filter dipakai: `?name=<name>`

| Field raw | Keterangan |
|---|---|
| `.id` | ID interface |
| `name` | Nama interface |
| `type` | Tipe (ether/vlan/bridge/dll) |
| `mtu` / `actual-mtu` / `l2mtu` | MTU |
| `mac-address` | MAC |
| `running` | Status link up |
| `disabled` | Status enable/disable |
| `rx-byte` / `tx-byte` | Total byte kumulatif |
| `rx-packet` / `tx-packet` | Total paket kumulatif |
| `comment` | Komentar |

### `/interface/monitor-traffic` (statistik real-time)
```
=interface=<interfaceName>
=once=
```
Field raw yang dikembalikan: `rx-bits-per-second`, `tx-bits-per-second`, `rx-packets-per-second`, `tx-packets-per-second`, `rx-drops-per-second`, `tx-drops-per-second`, `rx-errors-per-second`, `tx-errors-per-second`

**Field yang dibaca aplikasi:** `rx-bits-per-second`, `tx-bits-per-second`

### `/interface/enable` dan `/interface/disable` (dinamis via template string)
```
=.id=<id>
```
Command dibangun sebagai `` `/interface/${action}` `` dengan `action` = `enable` atau `disable`.

---

## 17. Ping (Diagnostik)

### `/ping`
```
=address=<host>
=count=<count>     // default '4'
```
Setiap balasan berisi field raw: `host`, `sent`, `received`, `packet-loss`, `min-rtt`, `avg-rtt`, `max-rtt`, `seq`, `time`, `ttl`

---

## 18. System — Resource, Identity, Clock, Log, Reboot

### `/system/resource/print`
| Field raw | Dipetakan aplikasi sebagai |
|---|---|
| `cpu-load` | `cpuLoad` |
| `cpu-count` | `cpuCount` |
| `cpu-frequency` | `cpuFrequency` |
| `free-memory` / `total-memory` | `memoryFree` / `totalMemory` (dikonversi ke MB) |
| `free-hdd-space` / `total-hdd-space` | `diskFree` / `totalDisk` (dikonversi ke MB) |
| `architecture-name` | `architecture` |
| `model` | `model` |
| `serial-number` | `serialNumber` |
| `firmware-type` | `firmware` |
| `voltage` / `board-voltage` | `voltage` (fallback) |
| `temperature` / `board-temperature` | `temperature` (fallback) |
| `bad-blocks` | `badBlocks` |
| `uptime` | `uptime` |
| `version` | `version` |
| `board-name` | `boardName` |

### `/system/identity/print`
Field raw: `name`

### `/system/identity/set`
```
=name=<name>
```

### `/system/clock/print`
Field raw: `time`, `date`, `time-zone-name`, `gmt-offset`

### `/system/reboot`
Tanpa parameter.

### `/log/print`
Filter dipakai: `?topics~<keyword>` (partial match)

| Field raw | Keterangan |
|---|---|
| `.id` | ID entry log |
| `time` | Waktu |
| `topics` | Kategori log |
| `message` | Isi pesan |

---

## Ringkasan Pola Umum di Codebase

1. **Lookup-then-mutate**: hampir semua `set`/`remove` didahului `print` dengan filter `?field=value` untuk mendapatkan `.id`, baru kemudian `=.id=<id>` dipakai pada `set`/`remove`.
2. **Konsistensi target ID**: mayoritas endpoint pakai `=.id=`, kecuali `/ip/hotspot/user/set` yang memakai `=numbers=` — inkonsistensi yang perlu diperhatikan bila mereplikasi pola ini ke modul lain.
3. **Field opsional dikirim kondisional**: banyak fungsi `add`/`set` (profile PPP, profile hotspot) membangun array param secara dinamis — field hanya di-push jika value ada, bukan selalu mengirim seluruh field sekaligus.
4. **Soft-suspend berlapis**: aplikasi ini punya 4 metode isolir independen — ganti PPP profile ke `isolir`, address-list + firewall drop rule, DHCP lease block, dan simple queue throttle — dipilih sesuai tipe koneksi pelanggan (PPPoE vs IP statik vs DHCP).