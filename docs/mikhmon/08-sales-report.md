# Modul 08 — Sales Report & Financial Engine

> Kembali ke [README](README.md) · Kode asli: `get/get_report.php` (per hari), `get/get_dashboard.php` `get_livereport` (per bulan), `view/report.php`, `view/live_report.php`.
>
> **Status implementasi di Polyglot: ✅ selesai (Fase 5)** — `ListReports` (filter legacy day/month/year + `summary_only`) dan `DeleteReport` sudah diekspos via `HotspotService`, **termasuk perbaikan bug filter gateway** (sebelumnya `?owner=mikhmon-report` → tidak pernah cocok dengan record asli).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| `get_report` (per hari) | `day` (contoh `aug/17/2026`), `f` | `/system/script/print ?source=<day>` (+ `count-only`) | jsEncode25 array |
| `get_livereport` (per bulan) | `day`, `month` (contoh `aug2026`), `f` | `/system/script/print ?owner=<month>` (+ `count-only`) | jsEncode25 array |

> Laporan **disimpan di dalam MikroTik** sebagai `/system/script` (bukan database) — dibuat oleh on-login script saat user login (modul 05 §4.1).

## 2. Format Record di RouterOS

Setiap login user voucher membuat `/system/script` dengan:
- **name** = `$date-|-$time-|-$user-|-<price>-|-$address-|-$mac-|-<validity>-|-<profile>-|-$comment`
- **owner** = `$month$year` → contoh `aug2026` (dipakai filter per bulan)
- **source** = `$date` → contoh `aug/17/2026` (dipakai filter per hari)
- **comment** = `mikhmon`

Contoh lengkap:
```
aug/17/2026-|-14:20:00-|-VIP123-|-3000-|-192.168.88.50-|-AA:BB:CC:DD:EE:FF-|-1h-|-1Jam-3K-|-vc-A1B-08.17.26-Event
```

> **PENTING:** delimiter pemisah adalah `-|-` (strip, pipe, strip). Nama profile (bagian 7) **tidak boleh mengandung `-|-`** — karena itu nama profile dinormalisasi spasi → `-` saat create (modul 05 §4.3).

## 3. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto` (message `HotspotReport` sudah ada); usecase: `internal/usecase/hotspot/hotspot_usecase.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go` (ListReports/DeleteReport); parser: `internal/driver/mikrotik/hotspot/report.go` (`ParseMikhmonTransactions`).

### 3.1 Get Sales Report — ✅ diekspos (`ListReports`)

`HotspotService/ListReports` — `ListHotspotReportsRequest{device_id, day, month, year, summary_only}` → `ListHotspotReportsResponse{reports[], total_income, total}`.

- **Filter (legacy):** `day` → `?source=<day>` (get_report), `month` (owner `aug2026`) → `?owner=<month>` (get_livereport), `year` → post-filter suffix tanggal. Tanpa filter → semua record Mikhmon (`?comment=mikhmon`) + `summary_only` (keputusan #3).
- **Bug yang diperbaiki:** gateway lama memakai `?owner=mikhmon-report` yang tidak pernah cocok dengan record asli (owner = `$month$year`, comment = `mikhmon`). Sekarang `ListReports(ctx, driver, port.ReportFilter)`.
- **Message `HotspotReport` (proto):** `{id, date, time, username, profile, price, comment}` — `price` proto `double`, dikonversi dari string di `mapper_report.go` (`ToProtoHotspotReports`, `SumReportIncome`).
- Handler: `report_handler.go`; usecase: `GetReports`/`GetReportsByFilter` mendelegasikan filter ke gateway.

### 3.2 Delete Report — ✅ diekspos (`DeleteReport`)

`HotspotService/DeleteReport` — `{device_id, ros_id}` → `/system/script/remove =.id=<id>` (baru, tidak ada di legacy).

## 4. Tipe Data (proto / port)

```protobuf
// api/proto/v1/hotspot.proto
message HotspotReport {
  string id = 1; string date = 2; string time = 3; string username = 4;
  string profile = 5; double price = 6; string comment = 7;
}
```

Tipe port: `port.MikhmonTransaction` — `{ros_id, raw_name, date, time, username, price, address, mac, validity, profile, comment}`. Parser aktual: `ParseMikhmonTransactions(result command.Result) []MikhmonTransaction` (memakai `mikrotik.ParseSystemScripts` + filter `strings.Contains(s.Name, "-|-")`).

## 5. Logika Khusus — Parser Record

Parser sudah diimplementasikan di `internal/driver/mikrotik/hotspot/report.go` — setara `ParseScriptReportName` pada draft awal, tetapi mengembalikan slice (memakai `mikrotik.ParseSystemScripts`) dan mempertahankan `RosID`/`RawName`:

```go
// ParseMikhmonTransactions: /system/script/print ?comment=mikhmon -> []MikhmonTransaction
// split name dengan "-|-": date, time, username, price, address, mac, validity, profile, comment
func ParseMikhmonTransactions(result command.Result) []MikhmonTransaction
```

## 6. Catatan Implementasi

1. **Filter ganda:** `GetReportsByFilter` → `gateway.ListReports(ctx, driver, port.ReportFilter{Day, Month, Year})` — `day`→`?source=`, `month`→`?owner=`, `year`→post-filter suffix; ketiganya kosong → semua record Mikhmon (keputusan #3, `summary_only` untuk efisiensi dashboard).
2. **Record rusak / tidak parseable:** `ParseMikhmonTransactions` melewati script tanpa `-|-`, tapi record partial (delimiter ada, segmen kurang) tetap di-parse; `SumReportIncome` hanya menjumlah `price` yang valid.
3. **Performa:** `summary_only` menghemat bandwidth dashboard; tambahkan `limit`/`offset` dan cache per (device, day/month) bila perlu.
4. **Timezone:** tanggal `aug/17/2026` berasal dari clock router; dokumentasikan bahwa laporan mengikuti timezone router, bukan server.
5. **Fitur lanjutan (opsional):** `GetTopSelling` — top profile/voucher terjual.
