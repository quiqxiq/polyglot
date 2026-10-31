# PLAN: Integrasi GenieACS TR-069 & OLT EPON HIOSO (ONU/OLT)

**Status:** Disetujui — implementasi berjalan
**Rujukan analisis:** `ANALISIS_FITUR_GLOBAL_SMARTAPP.md` §3.2 (OLT), §3.3 (ACS), §5.1 (linked ONU), §5.5 (PSB)
**Cakupan:** OLT HiOSO EPON (HA7302/HA7304), ONU/ONT lifecycle, ACS GenieACS (TR-069), integrasi PSB/PPPoE, frontend OLT/ONT.
**Di luar cakupan:** ODP/ODC/GIS, kalkulator redaman, remote ONT proxy, driver OLT Huawei/ZTE, server CWMP sendiri, web-scraping OLT.

---

## 1. Hasil Riset (terverifikasi)

### 1.1 HIOSO EPON OLT
- **Model:** HA7302/HA7302CS/HA7302CST (2 PON standalone, 64 ONU, maks 128, WEB/SNMP/CLI) dan HA7304 (4 PON).
- **SNMP v2c, enterprise `1.3.6.1.4.1.25355`.** OID kunci (teruji live HA7304/HA7302):
  - Nama ONU `25355.3.2.6.3.2.1.37.1.{idx}` (STRING)
  - MAC ONU `25355.3.2.6.3.2.1.11.1.{idx}`
  - Rx `25355.3.2.6.14.2.1.8.1.{idx}` (STRING dBm / `na`)
  - Tx `25355.3.2.6.14.2.1.4.1.{idx}`
  - Link-state `25355.3.2.6.3.2.1.39.1.{idx}` (1=Up, 2=Down) — **sumber kebenaran online**
  - Distance `25355.3.2.6.3.2.1.25.1.{idx}` (meter)
  - Firmware signature `25355.3.1.8.1.1.2.1` (deteksi varian, mis. `v7.76/HA7302CSM/…`)
  - OLT IP/subnet/gw `25355.3.1.1.0` / `.2.0` / `.3.0`
- **Index ONU:** HA7304 `{PON}.{ONU}` (slot 1); HA7302 LLID datar `1.{ONU}` (1..128).
- **Quirk SNMP:** timeout ≥5s & retries ≥2; walk **per-PON**; jangan pakai ifTable untuk status PON; `na` ≠ offline; roster carry-forward + `MAX_MISSED_POLLS=12` + anti-flap.
- **Telnet CLI:** Cisco-like, prompt `EPON>` → `EPON#`; CRLF; IAC negotiation (HA7302); login 3 lapis; banner panjang; paging `--More--`.
  - Read: `show onu info epon 0/{PON} all`, `show epon 0/{PON} optical-ddm`, `show version`, `show system`, `show running-config`, `show pon {N} statistic`.
  - Write HA7304: `interface epon 0/{PON}` → `onu {N} name {label}`; `onu {N} reboot`; `onu {N} activate|deactivate`; `delete onu {N}`; save `write`.
  - Write HA7302: `epon` → `pon 1/{PON}` → `set onu {N} reboot`; `delete onu 1/{PON}/{N}`; enable/disable via auth-mode deny/pass; **rename via SNMP SET** (butuh write community).
  - Larangan: `no onu …`, kirim `\n` saja, treat `na` sebagai offline.
- **Web UI** (GoAhead-Webs) bisa dibaca (`/onuConfigOnuList.asp?oltponno=0/1/1`) — hanya referensi field, **tidak** diimplementasikan.
- **Authorization flow ONU unconfigured** belum terdokumentasi publik (Hubsoft memakai Telnet dengan profil command). **Wajib spike Phase 0**; fallback: operator registrasi ONU di OLT, sistem hanya bind/rename.

### 1.2 GenieACS TR-069
- NBI REST: `GET /devices/?query=`, `POST /devices/{id}/tasks`, `PUT/DELETE /presets/{name}`, `GET /presets`, `DELETE /faults/{id}`.
- Task: `getParameterValues`, `refreshObject`, `setParameterValues`, `addObject`, `deleteObject`, `reboot`, `factoryReset`, `download`.
- NBI tanpa auth → wajib network-isolated; optional `x-api-key`.
- Bootstrap: preset set `ManagementServer.URL` + `PeriodicInformInterval=300`.
- Telemetri per model berbeda (TR-098 vs TR-181 vs VirtualParameters) → katalog path + fallback; Rx optik fallback ke OLT.
- Pola matching ACS↔ONU: serial → MAC (±1) → pin manual; konflik dilepas.

---

## 2. Keputusan Desain

| # | Keputusan |
|---|---|
| D1 | HIOSO-first EPON (HA7302/HA7304); vendor lain nanti di balik port yang sama. |
| D2 | Read = SNMP; Write = Telnet CLI. Web scraping tidak dipakai. |
| D3 | Deteksi varian otomatis (`sysObjectID` + firmware signature + `sysDescr`), override di `devices.extra["variant"]`. |
| D4 | Transport CLI: Telnet 23 default; SSH opsional (jika port 22 tersedia). |
| D5 | ONT identity = MAC (LLID) ternormalisasi, unik per OLT. |
| D6 | Status online dari link-state; `na`→NULL; anti-flap; roster carry-forward. |
| D7 | ACS = GenieACS NBI; server direpresentasikan sebagai `devices` tag `acs`. |
| D8 | Bootstrap ACS via preset idempoten (`PUT /presets/...`). |
| D9 | Telemetri via katalog parameter per ProductClass + fallback path. |
| D10 | Matching ACS↔ONT: pin manual → serial → MAC exact → MAC±1; konflik dilepas. |
| D11 | Semua aksi tulis via `router_jobs` (retry/backoff/DEAD/audit). |
| D12 | PSB mencatat OLT/PON/MAC/SN; aktivasi → job `ONT_BIND`. |
| D13 | Linked ONU tampil di subscription/customer; cek restart PPPoE memakai service eksisting. |
| D14 | Alert WA transisi + cooldown + dedupe. |
| D15 | Phase 0 spike wajib di perangkat nyata sebelum write commands final. |
| D16 | Multi-tenant `tenant-default`; kredensial di vault; telnet cleartext wajib VLAN/tunnel. |

---

## 3. Skema Database

### 3.1 `migrations/000036_olt_ont.up.sql`
- `registrations`: + `olt_device_id UUID REFERENCES devices(id)`, `pon_index INT`, `onu_index INT`.
- Tabel `onts` (kolom inti): id, tenant_id, olt_device_id, pon_index, onu_index, llid, mac_address, serial_number, model, firmware, state (`ONLINE/OFFLINE/UNKNOWN/REMOVED`), admin_state, rx_power_dbm, tx_power_dbm, distance_m, temperature_c, missed_polls, offline_strikes, rx_na_strikes, last_seen_at, last_polled_at, customer_id, subscription_id, pppoe_username, acs_device_id, acs_matched_by, acs_last_inform_at, acs_telemetry JSONB, wlan_slots, source, timestamps, unique (olt_device_id, mac_address), unique (olt_device_id, pon_index, onu_index).
- Tabel `ont_events`, `ont_metric_samples`.
- Seed settings: `olt.*` (timeout, retries, roster, strikes, poll intervals, retention, warning/critical Rx), `alert.ont_*`; template WA `ONT_OFFLINE`, `ONT_ONLINE`, `ONT_LOW_RX`.

### 3.2 `migrations/000037_acs.up.sql`
- `acs_devices`: id, tenant_id, acs_device_id UNIQUE, oui, product_class, serial_number, software_version, last_inform_at, tags, mac_candidates, raw JSONB, matched_ont_id, match_source, first_seen_at, last_seen_at.
- `acs_presets`: id, tenant_id, name UNIQUE, precondition JSONB, configurations JSONB, enabled, last_synced_at, last_error.
- Seed settings: `acs.*` (enabled, server_device_id, periodic inform, preset name, param catalog, cron).

---

## 4. Arsitektur

```
ConnectRPC OltService/OntService
  → usecase/olt, usecase/ont
      → port.ONTRepository (postgres)
      → port.OLTManager  ← adapter/oltmanager/hioso → registry.Get → *hioso.Driver
      → port.ACSManager  ← adapter/acsmanager/genieacs → genieacs.NBI
  → router_jobs worker (async write)
scheduler: olt-sync (5m), acs-match (15m), acs-telemetry (15m), retention (harian)
```

---

## 5. Rencana Fase

| Fase | Isi | Estimasi |
|---|---|---|
| A | Dokumentasi + migrasi 000036/000037 + domain/port/repo/model/fake + smoke test | 2–3 hari |
| B | Driver hioso SNMP (OID/detection/walk/parser) + adapter oltmanager + usecase sync/read + OltService + UI baca OLT/ONU | 1,5–2 minggu |
| C | Driver hioso Telnet/SSH + write commands + jobs + alert WA + grafik Rx | 1,5–2 minggu |
| D | genieacs NBI + acsmanager + usecase telemetry/matching/actions + OntService + UI ACS | 2–3 minggu |
| E | Integrasi PSB/registration + linked ONU + PPPoE + polish + dokumentasi | 1–2 minggu |

### Acceptance ringkas
- HA7302 & HA7304 terdeteksi; inventori ±1 poll; Rx benar; `na` tidak offline; anti-flap.
- Aksi rename/reboot/enable/delete/save berhasil di kedua varian; command tepat; job retry/DEAD; audit.
- Preset ACS membuat ONT inform 300s; telemetri minimal (Rx/temp/SSID/clients/SN/last inform); aksi remote sukses; konflik match dilepas.
- PSB → ONU rename & link; LOS/Rx tampil di pelanggan; PPPoE restart bekerja.
- Quality gates: build/vet/test/boundary/buf lint/codegen idempotent/migration smoke/tsc.

---

## 6. Risiko & Asumsi
- Flow authorization ONU unconfigured belum publik → spike; fallback manual.
- Variasi firmware/command → capability matrix + override.
- SNMP lossy → per-PON, roster, anti-flap.
- Telnet cleartext & sesi terbatas → VLAN/tunnel + mutex/advisory lock.
- ACS connection request gagal di CGNAT → periodic inform; task queued = retry.
- Path parameter ONT bervariasi → katalog + fallback OLT.
- MAC±1 heuristik → pin manual menang; konflik dilepas.

## 7. Referensi
- [spesifikasi HA7302CS](https://en.hioso.com/cy/dtl/Encsprodetail_1/Sierra-Leone-1253761807820066816.html)
- [HIOSO cloud config](https://en.hioso.com/Download_detail/8.html)
- [netmiko hioso](https://github.com/ktbyers/netmiko/blob/develop/netmiko/hioso/hioso.py)
- [Hubsoft × Hioso (Telnet)](https://hubsoft.io/integracoes/hioso/)
- [KusumaVisionNMS — panduan HIOSO](https://github.com/Masamune21-dev/KusumaVisionNMS/blob/main/docs/SMARTOLT_HIOSO_GUIDE.md)
- [KusumaVisionNMS — GenieACS↔ONU matching](https://github.com/Masamune21-dev/KusumaVisionNMS/blob/main/app/Console/Commands/GenieacsMatchOnuCommand.php)
- [Kim-puL/OLT_MGMT — OID map HIOSO](https://github.com/Kim-puL/OLT_MGMT)
- [REST API HA7302CST (referensi field web)](https://github.com/takedakimochi/http-rest-api-olt-hioso)
- [ISPbills OLT management](https://ispbills.com/docs/en/monitoring/olt_management/)
- [GenieACS NBI API](https://docs.genieacs.com/en/stable/api-reference.html)

---

## 8. Status Implementasi (selesai)

**Backend (Go)**
- Migrasi `000036_olt_ont` + `000037_acs` (onts, ont_events, ont_metric_samples, acs_devices, acs_presets, kolom PSB, settings `olt.*`/`acs.*`/`alert.ont*`, template WA); smoke test hijau.
- Domain `internal/domain/olt` (varian/capability/validasi) dan `internal/domain/ont` (state/anti-flap, MAC, Rx, katalog telemetri, matching ACS, waktu PSB).
- Driver `internal/driver/hioso`: SNMP (OID 25355, deteksi varian, walk per-PON, roster-ready) + Telnet (IAC/CRLF/menu/3-tier login/paging) + command per varian (rename HA7304 CLI / HA7302 SNMP SET, reboot, admin, delete, save) + parser; tidak dial perintah bila varian tidak dikenal.
- Driver `internal/driver/genieacs/nbi.go` dkk: NBI client (devices, tasks, presets, faults, tags, query).
- Adapter: `oltmanager` (provider → `*hioso.Driver`), `acsmanager` (GenieACS NBI → telemetri/preset/aksi).
- Usecase: `olt` (CRUD + probe, sync roster/anti-flap/event/sampel, enqueue job), `ont` (SyncDevices + matching, telemetri, link/unlink, pin/unpin, aksi, alert WA, PSB binder).
- API ConnectRPC: `OltService` (16 RPC) + `OntService` (12 RPC), RBAC `olt:read/manage`, `ont:read/manage`, policy seeds.
- Job queue: tipe `SYNC_OLT`, `ONT_RENAME/REBOOT/SET_ADMIN/DELETE`, `OLT_SAVE_CONFIG`, `ACS_*`, `ONT_BIND` + handler + worker wiring.
- Scheduler: `olt-sync` (5m), `acs-match` (15m), `acs-telemetry` (15m) — env `OLT_SYNC_CRON`, `ACS_MATCH_CRON`, `ACS_TELEMETRY_CRON`.
- Integrasi PSB: field `olt_device_id`/`pon_index`/`onu_index` di registrasi & proto, binder pasca-konversi (rename ONT ke kode pelanggan + link ke customer/subscription).
- Alert WA: `ONT_OFFLINE`, `ONT_ONLINE`, `ONT_LOW_RX` dengan cooldown + dedupe timeline.

**Frontend (web/)**
- `/olt` (list + CRUD + test koneksi + sync) dan `/olt/$oltId` (PON/ONU table, filter, Rx badge, aksi job, sheet detail: info/riwayat optik/timeline).
- Settings `/settings/acs` (config ACS, ensure preset, sync devices, tabel matching + pin/unpin).
- Tab ONT di detail pelanggan (kartu telemetri/aksi/link-unlink).
- Dialog instalasi PSB: pilih OLT/PON/ONU index + tombol "Cari ONU" via MAC/SN.
- i18n id+en, sidebar, route tree ter-regenerate.

**Gate yang sudah dijalankan**
- `go build ./...` ✅, `go vet ./...` ✅, `go test ./... -race -cover` ✅ (semua paket ok).
- Migration smoke testcontainers 000036–000037 ✅.
- `buf lint` ✅, codegen idempoten ✅ (hash api/gen & web/src/gen sama setelah generate ulang).
- Boundary checks (`check-connect-errors`, `check-layer-boundaries`) ✅.
- Frontend: `tsc -p tsconfig.app.json` hanya 3 error pre-existing (i18n-provider `replaceAll`, plans-columns `ServiceType`), eslint/prettier file tersentuh ✅; `vite build` masih gagal karena error pre-existing tersebut (di luar scope).
- `govulncheck`: 22 temuan seluruhnya pre-existing (Go stdlib 1.27.1, x/net, x/text, excelize) — dependency baru (gosnmp) tidak terdaftar.

**Known limitations / next steps**
1. **Phase 0 spike di perangkat nyata belum dijalankan** (tidak ada akses OLT pada sesi ini): command write (`onu ... name/reboot/activate`, `delete onu ...`) dan flow *authorization ONU unconfigured* perlu diverifikasi pada HA7302/HA7304; semua akses terjadi lewat capability matrix + override, dan parser CLI sudah toleran. Untuk HA7302 enable/disable memakai dialek `auth-mode` yang ditandai TODO verifikasi.
2. Transport SSH belum diimplementasikan (Telnet default; `cli_transport=ssh` mengembalikan error jelas).
3. Job monitor UI/polling status job belum ada (aksi menampilkan job_id).
4. Preset ACS default memakai path TR-098; ONT berbasis TR-181 butuh override katalog/preset per ProductClass.
5. `make security` gagal karena temuan pre-existing di atas (bukan dari perubahan ini).
