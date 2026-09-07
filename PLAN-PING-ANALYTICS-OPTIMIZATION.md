# Rencana Implementasi & Optimasi: High-Frequency Ping Telemetry & TimescaleDB Ingestion

Dokumen ini merinci rencana komprehensif untuk mengatasi masalah degradasi performa, lonjakan latensi, *database connection pool starvation*, dan potensi *hang* pada sistem setelah fitur analitik ping berjalan lebih dari 12 jam.

---

## 1. Ringkasan Eksekutif & Karakteristik Beban (Workload)

Pada sistem **Polyglot**, analitik ping berjalan secara kontinyu menggunakan fitur streaming command `/ping` MikroTik dengan interval 1 detik.

### Karakteristik Volume Data:
- **1 Router / Device:** 1 ping/detik = 60 ping/menit = 3.600 baris/jam = **86.400 baris/hari**.
- **5 Router / Devices:** 5 ping/detik = **432.000 baris/hari**.
- **10 Router / Devices:** 10 ping/detik = **864.000 baris/hari** (~26 juta baris/bulan).

Setelah sistem berjalan **12 jam**, 1 router saja telah menghasilkan **43.200 baris**. Dengan beberapa router, ratusan ribu baris data mentah masuk ke database PostgreSQL/TimescaleDB secara terus-menerus.

---

## 2. Analisis Mendalam Akar Masalah (Root Cause Analysis)

Berdasarkan audit menyeluruh terhadap kode backend Go ([internal/usecase/metrics/ping_stream_manager.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/metrics/ping_stream_manager.go)), repository PostgreSQL ([internal/adapter/postgres/metrics_repository.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/postgres/metrics_repository.go)), konfigurasi pool ([internal/adapter/postgres/store.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/postgres/store.go)), skema migrasi ([migrations/000022_create_device_ping_metrics_table.up.sql](file:///home/quixiq/projects/polyground/polyglot/migrations/000022_create_device_ping_metrics_table.up.sql)), dan container Docker ([deployments/docker-compose.yml](file:///home/quixiq/projects/polyground/polyglot/deployments/docker-compose.yml)), ditemukan **6 akar masalah utama**:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                            AKAR MASALAH BOTTLENECK 12+ JAM                                  │
├───────────────────────────────┬───────────────────────────────┬─────────────────────────────┤
│ 1. Micro-Batching Ingestion   │ 2. Connection Pool Starvation │ 3. Cache & Index Thrashing  │
│    • Buffer cuma 10 baris     │    • MaxOpenConns cuma 25     │    • Postgres RAM cuma 128MB│
│    • Flush tiap 3 detik       │    • Habis direbut transaksi  │    • B-Tree index membengkak│
│    • Ratusan TX/menit         │      ping -> HTTP web HANG    │    • Random Disk I/O tinggi │
├───────────────────────────────┼───────────────────────────────┼─────────────────────────────┤
│ 4. Tanpa Kompresi TimescaleDB │ 5. Full Scan Query Agregasi   │ 6. Tanpa Timeout Ingestion  │
│    • Seluruh baris uncompressed│   • On-the-fly math di 43K+  │    • Worker memblokir jika  │
│    • 10x boros RAM & Storage  │     baris mentah tanpa CAGG   │      DB busy -> channel full│
└───────────────────────────────┴───────────────────────────────┴─────────────────────────────┘
```

### 1. Pola Micro-Batching & Transaction Overhead
- Di `ping_stream_manager.go`:
  ```go
  buffer := make([]device.PingMetricPoint, 0, 10)
  flushTicker := time.NewTicker(3 * time.Second)
  if len(buffer) >= 10 { flush() }
  ```
- Batch data di-flush setiap **3 detik** atau ketika mencapai **10 baris**.
- Di GORM `CreateInBatches(models, 100)`, setiap batch dieksekusi dalam transaksi terpisah (`BEGIN ... INSERT ... COMMIT`).
- Untuk 5 router, terjadi **100 transaksi database per menit** hanya untuk data ping. Setiap transaksi memicu Write-Ahead Log (WAL) sync (`fsync`) ke disk dan penguncian tabel.

### 2. Connection Pool Starvation (Kapasitas Maksimal Hanya 25 Koneksi)
- Di `store.go`, kapasitas koneksi database dibatasi:
  ```go
  sqlDB.SetMaxOpenConns(25)
  sqlDB.SetMaxIdleConns(10)
  ```
- Ketika operasi insert ping mulai melambat karena tabel membesar, worker ping menahan koneksi lebih lama.
- Seluruh 25 koneksi database habis terserap oleh background worker ping dan scheduler.
- **Dampak:** Request HTTP/ConnectRPC dari website (login admin, query subscriber PPPoE, dashboard, monitoring) tidak kebagian koneksi dan mengalami **hang / connection timeout**.

### 3. PostgreSQL Cache & Index Thrashing (Default `shared_buffers` 128 MB)
- Di `deployments/docker-compose.yml`, container TimescaleDB dijalankan tanpa kustomisasi memory buffer (default PostgreSQL `shared_buffers` hanya **128 MB**).
- Pada awal running, indeks B-Tree `idx_device_ping_metrics_device_time (device_id, recorded_at DESC)` muat di RAM.
- Setelah 12 jam, volume data dan ukuran indeks melampaui 128 MB. Setiap batch insert baru terpaksa melakukan **Disk I/O Thrashing (Random Read/Write ke Disk)** untuk mengupdate daun indeks B-Tree. Akibatnya, kecepatan insert turun drastis secara eksponensial.

### 4. Ketiadaan TimescaleDB Columnar Compression
- Hypertable `device_ping_metrics` dibuat tanpa mengaktifkan kebijakan kompresi (`timescaledb.compress`).
- Data time-series disimpan dalam format *uncompressed heap tuple*, yang memakan memori ~10 kali lebih besar dibandingkan penyimpanan kolumnar (*Gorilla/Delta-of-delta* compression) bawaan TimescaleDB.

### 5. Query Agregasi Mentah (On-The-Fly) Tanpa Continuous Aggregates
- Fungsi `QueryPingMetrics` di `metrics_repository.go` melakukan perhitungan agregasi matematis langsung di atas data mentah:
  ```sql
  SELECT MIN(NULLIF(rtt_ms, 0)), AVG(NULLIF(rtt_ms, 0)), MAX(NULLIF(rtt_ms, 0)),
         COALESCE(SUM(sent), 0), COALESCE(SUM(received), 0), COUNT(*)
  FROM device_ping_metrics WHERE device_id = ? AND recorded_at >= ? AND recorded_at < ?
  ```
- Setelah 12 jam, query ini harus membaca 43.200 baris per router dari disk/cache. Jika ada user membuka halaman analitik router atau melakukan refresh, database mengalami lonjakan CPU tinggi dan menahan koneksi pool selama beberapa detik.

### 6. Ketiadaan Context Timeout pada Ingestion Worker
- Pemanggilan `SavePingMetricsBatch(ctx, buffer)` menggunakan `ctx` panjang milik stream router tanpa batas waktu per-operasi (`context.WithTimeout`).
- Jika database lambat, goroutine `consumeStream` terkunci menunggu insert. Buffer channel streaming go-routeros (`out chan command.Result, 1000`) menjadi penuh, mengunci goroutine `pump`, dan menghentikan pembacaan socket TCP MikroTik (*backpressure stall*).

---

## 3. Arsitektur Solusi Baru

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                           ARSITEKTUR OPTIMASI INGESTION & QUERY                             │
└─────────────────────────────────────────────────────────────────────────────────────────────┘

 [MikroTik Router] ──(1 ping/detik)──> [go-routeros Stream Pump]
                                              │
                                              ▼
                                 [Smart Batching Ingestion]
                                 • Buffer: 30 - 60 poin (15 - 30 detik)
                                 • Dedicated Timeout Context (3 detik)
                                 • Non-blocking Backoff saat DB Busy
                                              │
                                              ▼
                                [Repository Batch Writer]
                                 • Session(SkipDefaultTransaction: true)
                                 • Direct Multi-Row Ingestion
                                              │
                                              ▼
                        ┌───────────────────────────────────────────┐
                        │          TIMESCALEDB HYPERTABLE           │
                        │           (device_ping_metrics)           │
                        │                                           │
                        │ • Chunk Interval: 1 Hari (Bukan 7 Hari)   │
                        │ • Foreign Key Dihapus dari Hypertable     │
                        │ • RAM Shared Buffers: 512MB - 1GB         │
                        └───────┬───────────────────────────┬───────┘
                                │                           │
                 (Data > 2 jam) │                           │ (Otomatis per menit)
                                ▼                           ▼
                ┌───────────────────────────────┐ ┌─────────────────────────────────┐
                │   COMPRESSION POLICY          │ │ CONTINUOUS AGGREGATE VIEW       │
                │   • SegmentBy: device_id, target│ │ (device_ping_metrics_1m)      │
                │   • OrderBy: recorded_at DESC │ │                                 │
                │   • Kompresi Kolumnar: 90%    │ │ • 1 Jam Data = 60 Baris         │
                │     hemat RAM & Storage       │ │ • 24 Jam Data = 1.440 Baris     │
                └───────────────────────────────┘ └────────────────┬────────────────┘
                                                                   │
                                                                   ▼
                                                       [QueryPingMetrics API]
                                                       (Dashboard / Grafik UI Cepat)
```

---

## 4. Rencana Aksi Implementasi Bertahap

### Fase 1: Database Migration (TimescaleDB Compression & Continuous Aggregates)
**Target File:**
- `[NEW]` [migrations/000028_optimize_ping_metrics_timescale.up.sql](file:///home/quixiq/projects/polyground/polyglot/migrations/000028_optimize_ping_metrics_timescale.up.sql)
- `[NEW]` [migrations/000028_optimize_ping_metrics_timescale.down.sql](file:///home/quixiq/projects/polyground/polyglot/migrations/000028_optimize_ping_metrics_timescale.down.sql)

**Langkah:**
1. **Atur Chunk Interval Menjadi 1 Hari:**
   ```sql
   SELECT set_chunk_time_interval('device_ping_metrics', INTERVAL '1 day');
   ```
2. **Hapus Foreign Key Constraint pada Hypertable:**
   Menghapus `REFERENCES devices(id) ON DELETE CASCADE` pada `device_ping_metrics` agar tidak ada overhead lock referensial pada setiap insert. Integritas data dijamin pada level usecase/service dan cleanup cron.
3. **Aktifkan Kebijakan Kompresi Kolumnar:**
   ```sql
   ALTER TABLE device_ping_metrics SET (
       timescaledb.compress,
       timescaledb.compress_segmentby = 'device_id, target',
       timescaledb.compress_orderby = 'recorded_at DESC'
   );

   SELECT add_compression_policy('device_ping_metrics', INTERVAL '2 hours');
   ```
4. **Buat Materialized Continuous Aggregate View (`device_ping_metrics_1m`):**
   ```sql
   CREATE MATERIALIZED VIEW device_ping_metrics_1m
   WITH (timescaledb.continuous) AS
   SELECT 
       time_bucket('1 minute', recorded_at) AS bucket_time,
       device_id,
       target,
       AVG(NULLIF(rtt_ms, 0))::real AS avg_rtt_ms,
       MIN(NULLIF(rtt_ms, 0))::real AS min_rtt_ms,
       MAX(NULLIF(rtt_ms, 0))::real AS max_rtt_ms,
       SUM(sent)::integer AS sent,
       SUM(received)::integer AS received,
       COUNT(*)::integer AS sample_count
   FROM device_ping_metrics
   GROUP BY bucket_time, device_id, target;

   SELECT add_continuous_aggregate_policy('device_ping_metrics_1m',
       start_offset => INTERVAL '1 hour',
       end_offset => INTERVAL '1 minute',
       schedule_interval => INTERVAL '1 minute');
   ```

---

### Fase 2: Optimasi Ingestion Worker di Backend Go
**Target File:**
- `[MODIFY]` [internal/usecase/metrics/ping_stream_manager.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/metrics/ping_stream_manager.go)

**Langkah:**
1. **Perbesar Ukuran Batch & Interval Flush:**
   - Ubah kapasitas buffer dari `10` menjadi `60`.
   - Ubah `flushTicker` dari `3 * time.Second` menjadi `15 * time.Second` (atau saat buffer mencapai 30–60 item).
   - Mengurangi frekuensi interaksi database hingga **80–90%**.
2. **Pasang Context Timeout pada Ingestion:**
   - Gunakan `context.WithTimeout(context.Background(), 5*time.Second)` saat memanggil `SavePingMetricsBatch` agar worker tidak pernah hang jika database mengalami disk latency spike.
3. **Pencatatan Error & Degradation Handling:**
   - Jika `SavePingMetricsBatch` gagal atau timeout, catat warning log terstruktur dan buang batch lama untuk mencegah buffer bloat / OOM.

---

### Fase 3: Optimasi Repository & Query Layer
**Target File:**
- `[MODIFY]` [internal/adapter/postgres/metrics_repository.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/postgres/metrics_repository.go)

**Langkah:**
1. **Bypass Transaction Overhead di GORM:**
   ```go
   return r.db.WithContext(ctx).
       Session(&gorm.Session{SkipDefaultTransaction: true}).
       CreateInBatches(models, 100).Error
   ```
2. **Optimasi QueryGrafik ke Continuous Aggregate:**
   - Di fungsi `QueryPingMetrics`: jika `bucket` bernilai `1m`, `5m`, `15m`, `1h`, atau rentang waktu > 1 jam, arahkan query langsung ke view `device_ping_metrics_1m` alih-alih men-scan tabel mentah.
   - Query 24 jam yang tadinya men-scan 86.400 baris mentah per router menjadi hanya membaca **1.440 baris** terkompresi.

---

### Fase 4: Database Infrastructure & Connection Pool Tuning
**Target Files:**
- `[MODIFY]` [internal/adapter/postgres/store.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/postgres/store.go)
- `[MODIFY]` [deployments/docker-compose.yml](file:///home/quixiq/projects/polyground/polyglot/deployments/docker-compose.yml)

**Langkah:**
1. **Penyesuaian Connection Pool di Go:**
   - Naikkan `sqlDB.SetMaxOpenConns` dari `25` menjadi `50` (atau `75`).
   - Naikkan `sqlDB.SetMaxIdleConns` dari `10` menjadi `25`.
2. **Tuning Parameter PostgreSQL di Docker Compose:**
   ```yaml
   postgres:
     image: timescale/timescaledb:latest-pg16
     command: 
       - postgres
       - -c
       - shared_preload_libraries=timescaledb
       - -c
       - shared_buffers=512MB
       - -c
       - work_mem=16MB
       - -c
       - maintenance_work_mem=128MB
       - -c
       - max_connections=150
   ```

---

## 5. Matriks Perbandingan Sebelum vs Sesudah

| Metrik / Parameter | Sebelum Optimasi | Sesudah Optimasi | Dampak Performa |
| :--- | :--- | :--- | :--- |
| **Frekuensi Transaksi DB** | Tiap 3 detik per router (~100 TX/menit) | Tiap 15–30 detik (~10–20 TX/menit) | **Pengurangan beban WAL & IOPS hingga 85%** |
| **GORM Transaction Mode** | `BEGIN ... COMMIT` per batch | `SkipDefaultTransaction: true` | **Menghilangkan overhead lock transaksi** |
| **Kapasitas Connection Pool** | 25 koneksi (sering habis) | 50–75 koneksi | **Mencegah website hang / antrean HTTP** |
| **Format Penyimpanan > 2 Jam** | Raw Heap Tuple (100% size) | TimescaleDB Compressed Columnar | **Hemat storage & RAM cache 90%** |
| **PostgreSQL Buffer Memory** | 128 MB (sering cache thrashing) | 512 MB – 1 GB | **Mencegah disk random read/write thrashing** |
| **Query Range 24 Jam** | Scan 86.400 baris mentah on-the-fly | Scan 1.440 baris dari Continuous Aggregate | **Query time turun dari ~1.500ms ke < 25ms** |
| **Risiko Worker Hang** | Tidak ada timeout pada batch insert | Strict 5s timeout & safe drop | **Driver streaming tidak akan pernah stall** |

---

## 6. Rencana Verifikasi & Validasi

1. **Verifikasi Migrasi Database:**
   - Jalankan `make migrate-up` dan verifikasi keberadaan hypertable, compression settings, dan continuous aggregate di TimescaleDB:
     ```sql
     SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name = 'device_ping_metrics';
     SELECT * FROM timescaledb_information.compression_settings;
     SELECT * FROM timescaledb_information.continuous_aggregates;
     ```
2. **Uji Unit & Smoke Test:**
   - Jalankan unit test worker dan repository:
     ```bash
     go test -v ./internal/usecase/metrics/... ./internal/adapter/postgres/...
     ```
3. **Uji Beban (Long-Run Simulation):**
   - Jalankan stream ping dengan 3–5 router simulasi selama 1 jam.
   - Monitor penggunaan koneksi database via `sqlDB.Stats()`:
     - `OpenConnections` harus tetap stabil di bawah ambang batas (ideal < 20).
     - `WaitCount` (koneksi yang antre) harus bernilai `0`.
   - Monitor responsivitas endpoint HTTP/ConnectRPC publik untuk memastikan tidak ada degradasi.
4. **Audit Integritas Arsitektur:**
   - Jalankan boundary checks dan linting:
     ```bash
     make check-connect-errors check-layer-boundaries lint build
     ```
