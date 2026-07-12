# Referensi Lapangan — 5 Repo ISP RT/RW-Net vs Plan Polyglot

Dokumen ini merangkum analisis **source code langsung** (bukan README) dari 5 repo
billing/monitoring ISP RT/RW-Net Indonesia yang dipakai di produksi, sebagai basis
bukti untuk keputusan di `docs/plan-provisioning/`. Repo ada lokal di `refrensi/`.
Konvensi lintas-issue yang lahir dari sini terkodifikasi sebagai **K9–K15** di
`README.md` folder ini; file ini adalah jejak buktinya.

> Ini melengkapi `ANALISIS-PROVISIONING-REPO-REFERENSI.md` (root repo) yang
> menganalisis 4 repo lewat git clone; di sini ditambah `mikhmonv3` dan pembacaan
> ulang source secara menyeluruh (7 tema paralel).

## Repo yang Dianalisis

| Repo | Bahasa | Kekuatan untuk plan | File kunci dibaca |
|---|---|---|---|
| `billing-rtrw` | Node.js | OLT 8-vendor SNMP, GenieACS built-in ACS, isolir PPPoE hook | `services/oltService.js` (2197 baris), `onuProvisionService.js`, `services/mikrotikService.js` (1708), `config/genieacs*.js`, `OLT_OID_REFERENCE.md` |
| `gembok-bill` | Node.js | Isolir 4-metode, voucher pricing, pppoe-monitor, rxPowerMonitor | `config/mikrotik*.js`, `config/pppoe-monitor.js`, `config/genieacs-commands.js`, `migrations/*voucher*.sql`, `docs/*isolir*.md` |
| `gembok-simple` | PHP | Hotspot cookie & scheduler, on-login parsing, voucher order | `admin/hotspot-*.php`, `includes/mikrotik_api.php`, `api/genieacs.php`, `voucher-order.php` |
| `mikhmonv3` | PHP | Voucher lifecycle kanonik (disable/remove/expired/reset) | `process/*hotspotuser.php`, `hotspot/hotspotactive.php` |
| `mikhmon-agent` | PHP | Isolir script + scheduler per-user, voucher generator, GenieACS | `lib/VoucherGenerator.class.php`, `lib/MikrotikService.class.php`, `scripts/isolir_profile_*.txt`, `ppp/*.php` |

## Temuan Utama per Area (bukti → dampak ke plan)

### A. MikroTik PPPoE / Hotspot — pemutusan sesi (K9)
- **`/ppp secret set profile=` & `disabled=yes` TIDAK memutus sesi online.** Bukti:
  `billing-rtrw/services/mikrotikService.js` `setPppoeProfile()` memanggil
  `kickPppoeUser()` (loop `/ppp active … remove`) **hanya bila** profil berubah;
  `gembok-bill/config/mikrotik.js` set secret lalu `/ppp/active/print ?name` →
  `/ppp/active/remove`; hotspot `setHotspotUserDisabled` set disabled lalu loop
  `/ip/hotspot/active/remove`. → **K9**: kill sesi wajib jadi langkah kedua
  teraudit untuk PPPoE & hotspot; Issue 03/04/13.
- **Kill hotspot ambil hanya `.id` pertama (bug shared-users>1).** Bukti:
  `mikrotikKickHotspotUser()` break setelah satu sesi. → Issue 13 Task 5: loop
  SEMUA `.id`.
- **mac-cookie auto-relogin.** Bukti: `gembok-simple/admin/hotspot-cookies.php`
  mengelola `/ip hotspot cookie` sebagai objek kelas-satu (get/delete). → K9/K13:
  mode A wajib `/ip hotspot cookie remove` saat putus.
- **Rate-limit hidup di `/ppp profile`, secret hanya referensi nama.** Bukti:
  `gembok-simple` `pppoe-profile.php` set `rate-limit` di profile; `addSecret`
  hanya name/password/profile/service. → Validasi pemisahan Issue 02 vs 03.

### B. Isolir / Suspend (K11, K12)
- **Isolir = change_profile ke profil ISOLIR + infra firewall/NAT.** Bukti:
  `gembok-bill` `ensureBlockedCustomersSetup()` (rule `src-address-list=
  blocked_customers action=drop` sekali); `billing-rtrw`
  `setupIsolirFirewall()`+`ensurePppProfileIsolirAddressListHook()` (on-up
  `/ip firewall address-list add list=LIST_ISOLIR address=$remote-address`,
  on-down remove); `mikhmon-agent/scripts/isolir_profile_*.txt` (on-login isolir
  hapus scheduler + tulis comment). → **K11**.
- **Profil isolir per-pelanggan/plan + ensure-exist fail-safe.** Bukti:
  `billing-rtrw` `customers.isolir_profile || 'isolir'`; `gembok`
  `ensureIsolirProfile()` auto-create hanya bila nama persis `isolir`, else warn.
  → **K11** (per-device entitas, resume ke profil paket).
- **4 metode isolir static-IP.** Bukti: `gembok-bill/config/staticIPSuspension.js`
  (address_list | dhcp_block | bandwidth_limit | firewall_rule via setting). →
  Issue 04: address_list kanonik + metode lain sebagai config.
- **Perangkat dulu, DB status mengikuti; degradasi anggun.** Bukti:
  `billing-rtrw` "UPDATE MIKROTIK DULU sebelum update database"; isolir-lokal
  bila `router_id` NULL. → **K12**.
- **Auto-suspend trigger 2 model.** Bukti: `gembok` grace_period_days (default 7
  sejak overdue); `billing-rtrw` `customer.isolate_day` (tanggal tetap, default
  10), flag `auto_suspension`/`auto_isolate` per-pelanggan. → issue scheduler baru
  (lihat Open Questions), kolom di Issue 03.

### C. GenieACS / TR-069 (Issue 08, 09, 10)
- **Semua aksi = task NBI** `POST /devices/<id>/tasks {name}`: setParameterValues,
  refreshObject, reboot, factoryReset, getParameterValues; `?connection_request`
  untuk eksekusi segera. Bukti: `billing-rtrw/config/genieacs.js`,
  `mikhmon-agent/genieacs/lib/GenieACS.class.php`. → Issue 09.
- **WiFi/WAN/RX = multi-path shotgun.** Bukti: `genieacs-commands.js` set SSID ke
  `InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID` + `Device.WiFi.SSID.1`;
  password ke `KeyPassphrase`/`PreSharedKey.1.KeyPassphrase`/`…PreSharedKey`; RX
  power dicoba berurutan `WANPONInterfaceConfig.RXPower`, `X_ALU-COM_RxPower`,
  `Device.XPON.Interface.1.RxPower`, dst. → **K13**, Issue 09/10.
- **acs_devices cache PK=`_id`.** Bukti: built-in ACS `billing-rtrw` tabel
  `acs_devices(id,serial_number,manufacturer,product_class,oui,params JSON,tags
  JSON,last_inform)`, `acs_tasks(device_id,name,payload,status)`. → Issue 08 skema;
  Issue 01 (acs_tasks ≈ provisioning_sync_log). "Online" = turunan `last_inform`
  + threshold (5 menit), bukan kolom status.
- **Linking 3-strategi.** Bukti: resolve device by tag `{_tags}` → PPPoE username
  `$or` banyak path → serial; join OLT↔ACS via SN ternormalisasi
  (`replace(/[^a-z0-9]/gi,'').toUpperCase()`). → Issue 08.

### D. OLT / ONU (Issue 05, 06, 07, 10)
- **Katalog OID per-profile (model), belasan field, probe-detect.** Bukti:
  `oltService.js` `BRAND_PROFILES` map brand→array profile; tiap profile
  `status_table/name_table/sn_table/tx_power_table/rx_power_table/probe_oid/
  distance_table/distance_tenths_meter/offline_reason_table/unauth_sn_table/…`;
  `OLT_OID_REFERENCE.md` (Hioso, 306 baris) + cache OID pemenang ke
  `data/olt_oids/{id}.json`. → **K13**, Issue 07.
- **`online_values` beda per merk & EPON/GPON, bisa string.** Bukti: `ONLINE_VALUES`
  (hioso [1,3,4], zte [1,3,'working','online'], huawei [5,1,'active'], …),
  `getOnlineValues()` override GPON. → Issue 05/07/08.
- **RX/TX decode auto-scale, bukan divider tetap.** Bukti: `parseSignal()` (>500÷100,
  >50÷10), `computeRxDbm()` (toSigned16, buang 0/65535, kandidat skala per-brand).
  → **K13**, Issue 10.
- **Discovery ONU dual-path: SNMP + CLI.** Bukti: `fetchUnauthOnus()` walk
  `unauth_sn_table` (hanya ZTE C300/C600, Huawei, Fiberhome); `onuProvisionService`
  CLI `show gpon onu uncfg` (ZTE), `display ont autofind` (Huawei),
  `show onu unregister` — untuk merk tanpa OID unauth. → Issue 05.
- **Authorize ONU sequence + TR-069 push, shell prompt-driven.** Bukti:
  `onuProvisionService` `interface gpon-olt_<pon>` → `onu <id> type <t> sn <sn>` →
  `interface gpon-onu_<pon>:<id>` → `tcont/gemport/vlan` → `pon-onu-mng … service` →
  `tr069 acs url …` (4 varian syntax) → `write`; SSH shell interaktif (kirim saat
  prompt `#|>`, timeout per-sequence). → Issue 06.
- **Terminate ONU = `no onu <id>`/`ont delete` (destruktif).** → Issue 04:
  suspend/resume di MikroTik; hanya TERMINATE menyentuh OLT (ClassDestructive/HITL).

### E. Voucher & Hotspot lifecycle (Issue 13)
- **Expired = `limit-uptime` native.** Bukti: `mikhmonv3/process/
  removeexpiredhotspotuser.php` query `?limit-uptime=1s` → `/ip/hotspot/user/remove`.
- **Revoke batch pilih by status.** Bukti: `removehotspotuserbycomment.php`
  filter `?comment=<batch> ?uptime=00:00:00` — hanya hapus voucher **belum
  dipakai**. → Issue 13 revoke.
- **disable = set disabled=yes SAJA (tak kill).** Bukti: `disablehotspotuser.php`.
  → K9 (plan menambah kill = peningkatan, bukan tiruan).
- **from_login via on-login script + comment guard + orphan cleanup.** Bukti:
  `generateHotspotExpiryScript()` guard prefix comment (`vc`/`up`/kosong);
  `removehotspotuser.php` hapus `/system script` & `/system scheduler`
  bernama=username sebelum remove user. → **K14/K15**, Issue 13 Task 9/10.
- **Validity satuan jam (bukan hari).** Bukti: `gembok-bill` `voucher_pricing.
  duration` (in hours), `voucher_online_settings.duration_type` default 'hours'. →
  Issue 13 Task 2: pakai value+unit / seconds, bukan `_days`.
- **Config generate kode.** Bukti: `VoucherGenerator` `isUsernameExists()` cek ke
  ROUTER (`/ip/hotspot/user/print ?name`), regen hingga maxAttempts; setting
  username_length/char_type/prefix/password_same. → Issue 13 Task 1/6.
- **Pricing berjenjang + jual-online generate-after-payment.** Bukti:
  `voucher_pricing(customer_price,agent_price,commission_amount)`; order publik
  status pending→paid→voucher dibuat di webhook. → issue billing baru (Open Q).

## Yang Tetap Beda dari Referensi (keputusan sadar)

- **No polling untuk push-capable protocol.** Semua referensi memonitor PPPoE
  dengan `setInterval` `/ppp active print` + diff (30s). Polyglot memakai stream
  native **`/ppp/active/print follow`** (Issue 12) — di RouterOS API setiap tabel
  ber-`print` mendukung `follow`, jadi tabel active langsung mem-push record
  terstruktur (bukan `/log` topic-parsing). `follow` juga meng-emit state awal
  lebih dulu → snapshot inheren tanpa panggilan terpisah; disconnect = record
  `.dead=yes`. Reconciliation hanya jaring pengaman opsional (koreksi Issue 12).
- **Satu jalur audit.** Referensi mengeksekusi command langsung; Polyglot lewat
  `provisioning_sync_log → command_audit_log` (K4). Script router = deviasi yang
  direkonsiliasi (K15).
- **Scheduler harian tunggal**, bukan objek per-user bernama=username (hindari
  orphan) — lebih bersih dari referensi.
- **PostgreSQL + InfluxDB** (bukan SQLite; time-series bukan di RDBMS).
