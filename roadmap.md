# ROADMAP INTEGRASI FITUR MIKRODASH KE POLYGLOT

Dokumen ini merinci **peta jalan strategis (Roadmap)** untuk mengadopsi dan mengintegrasikan 12 kapabilitas utama dari **MikroDash** ke dalam ekosistem **Polyglot NetOps Engine**.

Integrasi ini mentransformasikan implementasi MikroDash (yang berbasis monolithic SQLite, RouterOS direct raw API, dan web socket tunggal) ke dalam arsitektur enterprise Polyglot: **Clean Architecture (Hexagonal)**, **CQRS**, **ConnectRPC / Protobuf v1**, **PostgreSQL 16 + TimescaleDB**, dan **React 19 + TypeScript**.

---

## 📌 DAFTAR ISI

1. [Prinsip Arsitektur & Aturan Integrasi](#1-prinsip-arsitektur--aturan-integrasi)
2. [Pemetaan 12 Pilar Fitur MikroDash](#2-pemetaan-12-pilar-fitur-mikrodash)
   - [Pilar 1: Network Topology & Autonomous L2/L3 Discovery](#pilar-1-network-topology--autonomous-l2l3-discovery)
   - [Pilar 2: Geo Engine, ASN Intelligence & Geographic Maps](#pilar-2-geo-engine-asn-intelligence--geographic-maps)
   - [Pilar 3: Live Traffic Flow & Conntrack Sankey Visualizer](#pilar-3-live-traffic-flow--conntrack-sankey-visualizer)
   - [Pilar 4: Wireless Spectrum & Floorplan Heatmap](#pilar-4-wireless-spectrum--floorplan-heatmap)
   - [Pilar 5: Fleet Config Management & Canary Deployment](#pilar-5-fleet-config-management--canary-deployment)
   - [Pilar 6: Zero-Touch Provisioning (ZTP) via Secure Tunnel](#pilar-6-zero-touch-provisioning-ztp-via-secure-tunnel)
   - [Pilar 7: Automated Fleet Backups & Visual Line Diff](#pilar-7-automated-fleet-backups--visual-line-diff)
   - [Pilar 8: AI Network Assistant / Copilot (Human-in-the-Loop)](#pilar-8-ai-network-assistant--copilot-human-in-the-loop)
   - [Pilar 9: Live Diagnostic Suite (Sniffer PCAP, Torch, Traceroute)](#pilar-9-live-diagnostic-suite-sniffer-pcap-torch-traceroute)
   - [Pilar 10: Event Alerting & Multi-Channel Dispatcher](#pilar-10-event-alerting--multi-channel-dispatcher)
   - [Pilar 11: Modular Interactive Dashboard Grid](#pilar-11-modular-interactive-dashboard-grid)
   - [Pilar 12: Multi-Tenant Hierarchical Sites, RBAC & OIDC SSO](#pilar-12-multi-tenant-hierarchical-sites-rbac--oidc-sso)
3. [Matriks Fase Eksekusi & Prioritas](#3-matriks-fase-eksekusi--prioritas)
4. [Standar Kualitas & Non-Negotiable Rules](#4-standar-kualitas--non-negotiable-rules)

---

## 1. Prinsip Arsitektur & Aturan Integrasi

Setiap fitur yang diadaptasi dari MikroDash harus mematuhi batas lapisan (*layer boundaries*) Polyglot:

```
┌─────────────────────────────────────────────────────────────┐
│                    Web UI (React 19 + Vite)                 │
└──────────────────────────────┬──────────────────────────────┘
                               │ ConnectRPC (Protobuf) / WebSockets
┌──────────────────────────────▼──────────────────────────────┐
│  Adapter Layer (ConnectRPC Handlers, HTTP, Repositories)    │
│  internal/adapter/connect/* | internal/adapter/http/*       │
└──────────────────────────────┬──────────────────────────────┘
                               │ Inbound Ports
┌──────────────────────────────▼──────────────────────────────┐
│  Usecase Layer (Business Logic, Orchestration)              │
│  internal/usecase/*                                         │
└──────────────┬──────────────────────────────┬───────────────┘
               │                              │ Outbound Ports
┌──────────────▼──────────────┐┌──────────────▼───────────────┐
│ Domain Models & Entities    ││ Outbound Driver Implementations│
│ internal/domain/*           ││ internal/driver/mikrotik/*   │
└─────────────────────────────┘└──────────────────────────────┘
```

- **Zero-Import Rule**: `internal/domain` bebas dari dependensi eksternal. `internal/driver/mikrotik` tidak boleh mengimpor `internal/adapter` atau `internal/usecase`.
- **Strict Credential Protection**: Password, WiFi pre-shared key, private key, dan secret SNMP tidak pernah disimpan atau ditampilkan kembali ke antarmuka pengguna (*write-only secrets*).
- **File Limit Constraint**: Setiap berkas Go produksi dijaga ketat di bawah 500 baris kode (`AGENTS.md`).

---

## 2. Pemetaan 12 Pilar Fitur MikroDash

### Pilar 1: Network Topology & Autonomous L2/L3 Discovery
Visualisasi peta graf topologi interkoneksi perangkat secara real-time dan lintas router (*cross-router fleet graph*).

- **Komponen MikroDash Asal**: `internal/topology/`, `internal/collect/topology.go`, `web/src/pages/topology.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/topology/`: Model node, link, port, neighbor, dan koordinat layout `(X, Y)`.
  - `internal/usecase/topology/`: Agregator penemuan topologi:
    - Layer 2: Penemuan tetangga via MNDP/CDP/LLDP (`/ip/neighbor`), tabel FDB MAC Switch/Bridge (`/interface/bridge/host`).
    - Layer 3 & Tunnel: Pemetaan tabel ARP (`/ip/arp`), DHCP leases, dan WireGuard peers.
    - Graph Correlator: Menemukan relasi antar router dalam database Polyglot untuk menggambar interkoneksi antar-cabang.
  - `internal/port/topology/`: Kontrak repositori penyimpanan posisi tata letak (`LayoutRepository`).
  - `internal/adapter/connect/topology/`: Endpoint RPC `GetTopologyGraph`, `StreamTopologyUpdates`, `SaveNodePositions`.
- **Fitur Frontend**:
  - Drag-and-drop canvas dengan engine force-directed layout.
  - Badge throughput live pada tiap tautan fisik/virtual (Rx/Tx rate langsung dari streaming driver).
  - Sanitasi koordinat persisten anti-corruption (`CleanPositions`).

---

### Pilar 2: Geo Engine, ASN Intelligence & Geographic Maps
Subsistem geolokasi terpadu untuk mendeteksi lokasi geografis lalu lintas IP dan router tanpa dependensi layanan cloud eksternal.

- **Komponen MikroDash Asal**: `internal/geo/`, `internal/geoplace/`, `internal/asn/`, `web/src/pages/connections-worldmap.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/geo/`: Model `GeoPoint`, `CityInfo`, `ASNInfo`, `GeoRoute`.
  - `internal/usecase/geo/`:
    - Embedded MMDB Reader: Parser lokal database MaxMind GeoLite2 City (.mmdb) tanpa latensi API luar.
    - ASN Resolver & Classifier: Database lokal pemetaan AS Number ke nama provider/operator dan kategori (CDN, ISP, Edu, Business).
    - City Locator & Distance Calculator: Perhitungan koordinat router dan jarak latensi geografis.
  - `internal/adapter/connect/geo/`: Service `ResolveIPGeo`, `ResolveASN`, `GetFleetGeoLocations`.
- **Fitur Frontend**:
  - **Connections World Map**: Menampilkan garis busur animasi (*geodesic arcs*) lalu lintas aktif keluar router ke seluruh dunia.
  - **Fleet World Map**: Sebaran seluruh router di peta dunia dengan indikator latensi, status online/offline, dan insiden site.
  - **Traceroute Map**: Visualisasi hop-by-hop traceroute pada peta dunia interaktif.

---

### Pilar 3: Live Traffic Flow & Conntrack Sankey Visualizer
Pemantauan dan analisis aliran trafik mendalam langsung dari tabel pelacakan koneksi (*connection tracking*).

- **Komponen MikroDash Asal**: `web/src/pages/connections-sankey.ts`, `connections.ts`, `internal/collect/connections.go`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/traffic/`: Model `ConntrackFlow`, `SankeyNode`, `TopTalkerEntry`.
  - `internal/usecase/traffic/`:
    - Conntrack Stream Ingestion: Mengalirkan data `/ip/firewall/connection` dengan filtering cerdas.
    - Flow Aggregator: Mengelompokkan aliran berdasarkan hirarki: *Interface Asal $\rightarrow$ Host Lokal $\rightarrow$ Layanan/Port $\rightarrow$ Gateway WAN $\rightarrow$ ASN / Negara Tujuan*.
    - Top Talkers Analyzer: Perhitungan real-time host dengan konsumsi bandwidth tertinggi.
  - `internal/adapter/connect/traffic/`: Service `StreamConnections`, `GetSankeyFlowData`, `GetTopTalkers`.
- **Fitur Frontend**:
  - Diagram visualisasi aliran Sankey dinamis.
  - Filter interaktif berdasarkan protokol, status NAT, flag koneksi, dan subnet.

---

### Pilar 4: Wireless Spectrum & Floorplan Heatmap
Manajemen WiFi canggih dengan visualisasi denah lokasi (*floorplan heatmap*) dan pemantau spektrum.

- **Komponen MikroDash Asal**: `web/src/pages/wifi-map.ts`, `wireless-fa.ts`, `capsman.ts`, `internal/wifiscan/`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/wifi/`: Model `Floorplan`, `AccessPointMarker`, `SignalSample`, `FrequencySpectrum`.
  - `internal/usecase/wifi/`:
    - Heatmap Signal Calculator: Estimasi propagasi sinyal RF berdasarkan lokasi AP, daya pancar (Tx power), dan redaman dinding.
    - Spectrum Analysis Processor: Pemrosesan data `/interface/wifi/frequency-scan` untuk deteksi interferensi channel 2.4/5/6 GHz.
    - Dual-Stack Orchestrator: Abstraksi pengelolaan AP untuk RouterOS v6/v7 CAPsMAN legacy dan RouterOS v7 WiFi-Qcom/AX.
    - Roaming Tracker: Pelacakan riwayat perpindahan AP per MAC klien dan degradasi SNR/CCQ.
  - `internal/port/wifi/`: Kontrak penyimpanan file floorplan (`FloorplanStoragePort`).
  - `internal/adapter/connect/wifi/`: Service RPC `UploadFloorplan`, `GetHeatmapData`, `RunFrequencyScan`.

---

### Pilar 5: Fleet Config Management & Canary Deployment
Distribusi konfigurasi RouterOS massal yang aman dengan proteksi pemulihan otomatis (*fail-safe canary rollout*).

- **Komponen MikroDash Asal**: `internal/cfgtpl/`, `internal/cfgdeploy/`, `web/src/pages/config-management.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/cfgdeploy/`: Model `ConfigTemplate`, `DeploymentJob`, `CanaryTarget`, `DriftReport`.
  - `internal/usecase/cfgdeploy/`:
    - Pre-flight Syntax Checker: Validasi sintaks CLI RouterOS sebelum pengiriman.
    - Automatic Safe Restore Point: Pengambilan backup cepat otomatis sesaat sebelum eksekusi dimulai.
    - Phased Canary Pipeline: Eksekusi script ke 1 router percontohan terlebih dahulu, memverifikasi ketersediaan koneksi, baru melanjutkan ke seluruh armada.
    - Auto-Revert on Disconnect: Pembatalan otomatis (*rollback*) bila script memutuskan koneksi manajemen API.
    - Configuration Drift Detector: Membandingkan konfigurasi live terhadap template acuan.
  - `internal/adapter/connect/cfgdeploy/`: Endpoint `DeployTemplate`, `ValidateTemplate`, `InspectDrift`.

---

### Pilar 6: Zero-Touch Provisioning (ZTP) via Secure Tunnel
Otomasi pendaftaran router baru dari pabrik/lapangan langsung ke sistem tanpa konfigurasi manual awal.

- **Komponen MikroDash Asal**: `internal/ztp/`, `web/src/pages/ztp.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/ztp/`: Model `ProvisioningToken`, `PendingDevice`, `ZTPScriptConfig`.
  - `internal/usecase/ztp/`:
    - Bootstrap Script Generator: Membuat script `.rsc` one-time yang mengonfigurasi IP sementara, WireGuard tunnel, dan pembatasan firewall API.
    - Dial-Home Ingestion: Menerima koneksi panggilan awal router dari balik NAT publik melalui userspace WireGuard endpoint.
    - Staging & Approval Queue: Menempatkan router di antrean verifikasi sampai disetujui admin untuk dimasukkan ke site produksi.
  - `internal/adapter/connect/ztp/`: Service `GenerateZTPScript`, `ListPendingDevices`, `ApproveDevice`.

---

### Pilar 7: Automated Fleet Backups & Visual Line Diff
Sistem pencadangan armada router terjadwal, efisien, dengan visualisasi perbandingan teks seperti Git.

- **Komponen MikroDash Asal**: `internal/backups/`, `web/src/pages/backups.ts`, `diffview.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/backup/`: Model `BackupRecord`, `ConfigDiffResult`, `RetentionPolicy`.
  - `internal/usecase/backup/`:
    - Dual-Format Backup: Pengambilan otomatis arsip biner `/system/backup` dan ekspor teks `/export`.
    - Change-Detection Hash: Hanya menyimpan file jika hash konfigurasi berubah dari backup sebelumnya (menghemat storage).
    - Unified Line Diff Engine: Membandingkan 2 backup teks dan menghasilkan diff baris (*added, deleted, modified*).
    - Migration Export Cleaner: Ekspor konfigurasi terenkripsi yang membersihkan MAC address spesifik agar siap di-restore ke perangkat dengan model/arsitektur berbeda.
  - `internal/port/backup/`: Kontrak penyimpanan arsip enkripsi (`BackupStoragePort`).
  - `internal/adapter/connect/backup/`: Service `TriggerBackup`, `GetBackupDiff`, `DownloadEncryptedBackup`.

---

### Pilar 8: AI Network Assistant / Copilot (Human-in-the-Loop)
Asisten cerdas pendiagnosis dan operator jaringan dengan persetujuan manusia (*human-in-the-loop*).

- **Komponen MikroDash Asal**: `internal/aicontext/`, `internal/aitools/`, `internal/aiprovider/`, `web/src/pages/ai-agent.ts`.
- **Target Arsitektur Polyglot**:
  - Integrasi dengan modul `internal/domain/llm/` dan `internal/usecase/llm/` yang sudah ada di Polyglot.
  - Kemampuan Baru:
    - Multi-Provider Support: OpenAI, Anthropic Claude, model lokal via Ollama / vLLM / LiteLLM.
    - Network Tool Calling: Tool khusus membaca metrik live, log error, tabel routing, dan hasil audit `secscan`.
    - Safe Execution Plan: AI menyusun proposal rencana tindakan (*Action Plan*), admin wajib menekan tombol persetujuan (*Approve*) sebelum dieksekusi di router.
    - Granular Undo Log: Setiap perubahan yang dilakukan AI dicatat dalam riwayat dengan tombol *Undo* otomatis.

---

### Pilar 9: Live Diagnostic Suite (Sniffer PCAP, Torch, Traceroute)
Perangkat diagnostik tingkat lanjut langsung dari browser untuk investigasi masalah jaringan.

- **Komponen MikroDash Asal**: `internal/diag/`, `web/src/pages/tools.ts`, `tools-sniffer-cards.ts`, `tools-torch-cards.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/diag/`: Model `PingSession`, `TracerouteHop`, `TorchFlow`, `PacketCaptureConfig`.
  - `internal/usecase/diag/`:
    - Web Packet Sniffer: Mengontrol `/tool/sniffer` RouterOS dan streaming hasil capture ke file `.pcap` yang siap diunduh dan dianalisis di Wireshark.
    - Web Torch: Streaming aliran paket real-time per interface (IP asal/tujuan, port, protokol, Tx/Rx rate).
    - Latency Jitter Profiler: Ping interaktif dengan visualisasi variasi latensi (jitter) dan packet loss.
    - Bandwidth Speed Tester: Pengujian throughput TCP/UDP antar router MikroTik di jaringan.
  - `internal/adapter/connect/diag/`: Service `StartTorchStream`, `StartSniffer`, `DownloadPCAP`, `RunBandwidthTest`.

---

### Pilar 10: Event Alerting & Multi-Channel Dispatcher
Engine peringatan insiden proaktif dengan pengiriman multi-saluran ke berbagai platform pesan instan.

- **Komponen MikroDash Asal**: `internal/alert/`, `internal/alertdispatch/`, `internal/notify/`, `web/src/pages/notifications.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/notification/`: Perluasan model `AlertRule`, `IncidentEvent`, `ChannelConfig`.
  - `internal/usecase/notification/`:
    - Metric Threshold Monitor: Evaluasi berkala atas CPU spike, memori rendah, interface flap/down, link degradasi, packet loss ping, dan BGP peer drop.
    - Cooldown & Flap Damping: Mencegah banjir notifikasi jika interface berulang kali hidup-mati dalam waktu singkat.
    - Multi-Channel Dispatcher: Driver pengiriman pesan ke:
      - Telegram Bot
      - Discord Webhook
      - Slack Webhook
      - Email (SMTP)
      - Pushover, ntfy, Gotify, Pushbullet, dan Custom HTTP Webhook JSON.
  - `internal/adapter/connect/notification/`: Service `ConfigureAlertRules`, `TestNotificationChannel`.

---

### Pilar 11: Modular Interactive Dashboard Grid
Dashboard terpadu modular berbasis grid interaktif yang dapat disesuaikan oleh tiap pengguna.

- **Komponen MikroDash Asal**: `web/src/pages/dashboard.ts`, `dashboard-grid.ts`, `dashcards/`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/setting/`: Model `DashboardLayoutConfig`, `WidgetPosition`.
  - `internal/usecase/setting/`: Penyimpanan dan sinkronisasi preferensi tata letak per-user ke PostgreSQL.
  - **Katalog Widget**:
    - *RJ45/SFP Physical Port Matrix*: Mockup port fisik router dengan indikator warna status link (Down, 100M, 1G, 10G, SFP).
    - *Hardware Health Gauges*: CPU load, memori RAM terpakai, ruang disk, temperatur, dan voltase.
    - *Live WAN Traffic Sparkline*: Aliran bandwidth realtime dengan riwayat 60 detik.
    - *Mini Topology & Security Scan Score*: Ringkasan skor kepatuhan keamanan dan status node utama.
  - `internal/adapter/connect/setting/`: Endpoint RPC `GetDashboardLayout`, `SaveDashboardLayout`.

---

### Pilar 12: Multi-Tenant Hierarchical Sites, RBAC & OIDC SSO
Pemisahan hak akses multi-cabang terstruktur dan integrasi login tunggal tingkat enterprise.

- **Komponen MikroDash Asal**: `internal/sites/`, `internal/rbac/`, `internal/oidc/`, `web/src/pages/settings-sso.ts`.
- **Target Arsitektur Polyglot**:
  - `internal/domain/tenant/`: Model `SiteHierarchy` (*Fleet $\rightarrow$ Site/Cabang $\rightarrow$ Area/Lantai $\rightarrow$ Router*).
  - `internal/usecase/auth/`:
    - Scope-based Access Control: Izin baca (*Read*) dan tulis (*Write*) yang dapat dibatasi pada level spesifik: global, per-site, atau per-router individual.
    - OpenID Connect (OIDC) Authenticator: Integrasi login SSO dengan Keycloak, Authentik, Okta, Google Workspace, atau Microsoft Entra ID.
    - Role Claim Mapping: Pemetaan otomatis grup OIDC dari provider eksternal ke peran (*roles*) internal Polyglot.
  - `internal/adapter/connect/auth/`: Handler autentikasi dan otorisasi sesi ConnectRPC.

---

## 3. Matriks Fase Eksekusi & Prioritas

Roadmap implementasi dibagi ke dalam 4 gelombang rilis (*Release Waves*) terukur:

| Wave | Fokus Utama | Pilar Terkait | Estimasi Kompleksitas | Target Deliverables |
| :--- | :--- | :--- | :--- | :--- |
| **Wave 1** | **Visualisasi Jaringan & Geografis** | Pilar 1 (Topology), Pilar 2 (Geo & ASN), Pilar 3 (Traffic Flow) | Medium-High | - Domain & Usecase Topology<br>- Engine MaxMind MMDB & ASN<br>- ConnectRPC Streaming Service<br>- Canvas Topologi & World Map |
| **Wave 2** | **Operasional & Diagnostik Live** | Pilar 9 (Diagnostics), Pilar 7 (Backups & Diff), Pilar 11 (Dashboard Grid) | Medium | - Live Packet Sniffer & Torch<br>- Backup Scheduler & Git Diff<br>- Widget Port RJ45/SFP Grid |
| **Wave 3** | **Manajemen Armada & Otomasi** | Pilar 5 (Config Management), Pilar 6 (ZTP), Pilar 4 (WiFi Spectrum) | High | - Canary Rollout Pipeline<br>- ZTP WireGuard Ingestion<br>- WiFi Floorplan Heatmap |
| **Wave 4** | **AI Copilot, Notifikasi & Enterprise** | Pilar 8 (AI Assistant), Pilar 10 (Alerting), Pilar 12 (Sites & OIDC) | Medium-High | - AI Tool Calling & Plan Approval<br>- Dispatcher Multi-Channel<br>- OIDC SSO & Scoped RBAC |

---

## 4. Standar Kualitas & Non-Negotiable Rules

Setiap kode yang dihasilkan dalam implementasi roadmap ini wajib melewati seluruh Quality Gates Polyglot:

1. **Unit Testing & Data Race Free**:
   - Seluruh kode domain, usecase, dan driver wajib memiliki table-driven tests.
   - Wajib lolos `go test -race ./...`.
2. **Boundary Checks**:
   - `make check-layer-boundaries` (Arsitektur heksagonal bersih).
   - `make check-connect-errors` (Error transport terstandarisasi).
3. **Kepatuhan Ukuran Berkas**:
   - Berkas produksi Go tidak boleh melebihi 500 baris kode (`AGENTS.md`). Modul besar wajib dipecah berdasarkan sub-fitur.
4. **Keamanan Kredensial (Zero-Leakage Policy)**:
   - Nilai rahasia (password, pre-shared key, private key) tidak boleh diekspos ke logging, WebSocket payload, atau response API biasa.
