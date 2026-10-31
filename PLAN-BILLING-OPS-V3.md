# PLAN-BILLING-OPS-V3 — Operasional Billing Setara Global SmartApp

Status: **SELESAI (W0–W8)** — diimplementasikan pada 2026-10-31.
Dokumen riset acuan: [`RISET-GLOBAL-SMARTAPP-LIVE.md`](RISET-GLOBAL-SMARTAPP-LIVE.md).

## 1. Tujuan

Membuat operasional billing/keuangan Polyglot mampu menjalankan alur kerja
seperti Global SmartApp: tagihan otomatis siklus, tagihan manual, PPN, master
metode pembayaran, pembayaran massal, reversal, sinkronisasi gateway &
settlement, hutang pelanggan, penagihan, laporan keuangan, dan ekspor.

## 2. Keputusan Bisnis (LOCKED)

1. **TIDAK ADA denda keterlambatan.** Keterlambatan ditangani lewat pengingat
   (pre/post jatuh tempo), laporan umur piutang, dan isolir.
2. **TIDAK ADA biaya tambahan/extra fee** pada tagihan otomatis maupun manual.
   Total tagihan = harga paket/prorata − diskon + PPN. Tidak ada baris biaya
   admin, biaya platform, atau fee siklus.
3. PPN: tarif paket (`service_plans.tax_percent`) menang; fallback
   `isp.tax_percent`. Mode `EXCLUSIVE` (harga + PPN) atau `INCLUSIVE`
   (PPN di dalam harga).
4. Mode billing default `PASCABAYAR`; `PRABAYAR` menerbitkan invoice
   `isp.prepay_invoice_lead_days` hari sebelum siklus dimulai.
5. `isp.closing_date` (1–28) menjadi anchor default langganan baru bila
   `billing_day`/`billing_anchor_day` kosong. Tidak ada period hard-lock.
6. Status "Belum Bayar H-N" hanya kosmetik/filter (`isp.unpaid_days_display`).
7. Uang tetap `float64` dengan pembulatan 2 desimal di domain; single tenant
   default `tenant-default`.
8. Hutang pelanggan bersifat manual (bukan denda). Customer dengan hutang
   tidak bisa dihapus (FK `ON DELETE RESTRICT`).
9. Export dibatasi maksimal 50.000 baris (invoice), 10.000 (hutang).

## 3. Migrasi Database

| Migrasi | Isi |
| --- | --- |
| `000038_billing_ops_v3` | `invoices.tax_percent`, sequence `invoice_no_seq`, settings mode billing/closing/PPN/reminder, template `INVOICE_CREATED`, `BILL_OVERDUE` |
| `000039_debts` | tabel `customer_debts`, `debt_payments`, kategori kas `cc-debt-income`, settings hutang, template `DEBT_REMINDER`, `DEBT_PAID` |
| `000040_payment_ops` | kolom rekening/QRIS/instruksi `payment_methods`; tabel `gateway_settlements`, `bulk_payment_batches`, `bulk_payment_items`; kategori `cc-gateway-fee`; template `PAYMENT_REJECTED` |

Smoke test migrasi (`internal/adapter/postgres/migrations_smoke_test.go`)
diperbarui: 46 key `isp.*`, 16 template, 7 kategori kas, 3 sequence, 5 tabel
baru, 7 kolom baru.

## 4. Fase Implementasi

### W0 — Fondasi
- Migrasi 000038–000040, domain `billing`/`debt`, port repository baru,
  model GORM + converter, seed RBAC role `finance`, mapping permission RPC.
- Anti-regresi: tidak ada objek denda/extra fee di migrasi maupun kode.

### W1 — Mode Billing, Tutup Buku, Tampilan Belum Bayar
- `domain/setting.ISPSettings` + `port.LoadISPSettings` membaca seluruh setting
  baru (mode, closing, lead days, PPN, offset reminder, sync gateway, hutang,
  mass payment).
- `usecase/billing/run_billing.go`: anchor default dari closing date, mode
  `PRABAYAR`/`PASCABAYAR`, jatuh tempo prabayar = awal siklus, PPN
  exclusive/inclusive + fallback setting.
- Scheduler menjalankan periode berjalan **dan** periode depan (prabayar),
  pascabayar otomatis skip periode depan.
- Registrasi PSB memakai `isp.closing_date` sebagai anchor default.

### W2 — Tagihan Manual + PPN + Filter/Summary/Export
- `InvoiceUseCase.CreateManualInvoice` (tipe `MANUAL`, item AD_HOC,
  nomor invoice, kode bayar, QR payload, notifikasi `INVOICE_CREATED`).
- `InvoiceRepository.FindFiltered/Summary/FindOverdueCandidates`
  (status, periode, tipe, rentang jatuh tempo, overdue/unpaid, search nama
  pelanggan, pagination + total).
- Export invoice CSV/XLSX (maks 50.000 baris).
- Frontend: filter tagihan server-side, ringkasan total, dialog "Tagihan
  Manual", tombol Ekspor.

### W3 — Metode Pembayaran, Mass Payment, Reversal
- Master `payment_methods`: CRUD + `SetActive`, kolom bank/QRIS/instruksi,
  seed awal tetap 4 metode.
- `BulkPaymentRepository` + `MassPaymentUseCase`: batch + item QUEUED/
  PROCESSING/SUCCEEDED/FAILED, claim atomik `FOR UPDATE SKIP LOCKED`, job
  router `MASS_PAYMENT`.
- Reversal pembayaran kini me-rollback `paid_through`, `end_date`,
  `next_due_date` langganan (sebelumnya tidak).
- Frontend: halaman pengaturan Metode Pembayaran.

### W4 — Operasional Gateway
- Status `SETTLED` bersifat terminal: webhook/sinkronisasi tidak boleh
  menurunkan status (anti-downgrade).
- `SyncPendingTransactions` + scheduler/manual RPC `SyncGatewayTransactions`.
- Jurnal beban biaya gateway (`cc-gateway-fee`) saat transaksi settled.
- Tabel `gateway_settlements` + repository + RPC list/save.
- Frontend: bagian laporan/operasional gateway (settlement) tersedia lewat
  halaman keuangan; hook API tersedia untuk lanjutan UI khusus.

### W5 — Hutang Pelanggan
- Domain + repository atomik `ApplyPayment`: validasi overpay, insert
  `debt_payments`, update sisa/status, jurnal kas masuk (`DEBT_PAYMENT`),
  semua dalam satu transaksi.
- `usecase/debt.Service`: create/update/cancel (guard pembayaran), pay,
  summary, list payments, export CSV/XLSX, notifikasi `DEBT_PAID`.
- Proto `DebtService` + Connect handler + route frontend `/debts`.
- Frontend: halaman daftar hutang, dialog catat hutang & terima pembayaran,
  ringkasan hutang.

### W6 — Collection & Dunning (Tanpa Denda)
- `ReminderWorker`: pengingat pre-due (`isp.reminder_pre_offsets`) dan
  post-due (`isp.reminder_post_offsets`, template `BILL_OVERDUE`), idempoten
  per invoice per hari.
- `usecase/collection.Service`: ringkasan piutang, daftar pelanggan overdue,
  kirim pengingat massal, isolir/restore massal. **Tidak ada denda.**
- Frontend: laporan rincian piutang per pelanggan + filter tagihan overdue.

### W7 — Laporan & Ekspor
- `AdvancedReportRepository` + `ManageReportUseCase`: `TaxReport`,
  `SalesReport`, `CollectionReport`, `CashflowReport`, `AgingDetail`,
  `ClosingReport`, `ExportReport` (CSV/XLSX).
- Proto `ReportService` diperluas; handler + router Connect.
- Frontend: tab Pajak PPN, Penjualan, Penagihan, Arus Kas, Rincian Piutang,
  Tutup Buku + tombol ekspor pada halaman laporan keuangan.

### W8 — Hardening & Gate
- Build/vet/test backend hijau; test baru: run billing v3 (prabayar, PPN
  inclusive, fallback PPN, closing anchor), reminder post-due, hutang
  (partial/overpay/cancel/export), tagihan manual, repository PostgreSQL
  (payment method, hutang+jurnal, batch claim, settlement, invoice filter).
- Frontend: `tsc -p tsconfig.app.json --noEmit` bersih, `vite build` sukses,
  ESLint file tersentuh bersih.
- Boundary check ConnectRPC & layer arsitektur lulus; `buf lint` +
  `buf generate` idempoten.

## 5. Permukaan API Baru

### BillingService
`CreateInvoice`, `ExportInvoices`, `ListPaymentMethods`, `SavePaymentMethod`,
`DeletePaymentMethod`, `CreateMassPayment`, `ListMassPayments`,
`GetMassPayment`, `ListGatewayTransactions`, `SyncGatewayTransactions`,
`ListGatewaySettlements`, `SaveGatewaySettlement`.
`ListInvoices` kini menerima filter lengkap + mengembalikan `summary` & `total`.

### DebtService
`ListDebts`, `GetDebt`, `CreateDebt`, `UpdateDebt`, `CancelDebt`, `PayDebt`,
`ListDebtPayments`, `GetDebtSummary`, `ExportDebts`.

### ReportService
`TaxReport`, `SalesReport`, `CollectionReport`, `CashflowReport`,
`AgingDetail`, `ClosingReport`, `ExportReport`.

## 6. Settings Baru (`system_settings`)

`isp.billing_mode`, `isp.closing_date`, `isp.unpaid_days_display`,
`isp.prepay_invoice_lead_days`, `isp.tax_percent`,
`isp.reminder_pre_offsets`, `isp.reminder_post_offsets`,
`isp.gateway_sync_stale_minutes`, `isp.debt_income_category_id`,
`isp.debt_reminder_enabled`, `isp.debt_reminder_days`,
`isp.gateway_fee_expense_category_id`, `isp.mass_payment_enabled`.

## 7. RBAC

Role baru `finance`: `billing:.*`, `debt:.*`, `cashbook:.*`, `report:.*`,
`customer:read`, `subscription:read`, `notification:read`.
`admin` mendapat `debt:.*`; `agent` mendapat `debt:read`.
Permission RPC baru dipetakan di `internal/adapter/auth/procedure_permissions.go`.

## 8. Frontend Baru/Diubah

- `/debts` — hutang pelanggan (daftar, filter, catat, bayar, ekspor).
- `/settings/billing` — mode siklus, tutup buku, PPN, reminder, mass payment,
  operasional gateway.
- `/settings/payment-methods` — CRUD master metode pembayaran.
- `/invoices` — filter server-side, ringkasan, tagihan manual, ekspor.
- `/reports/finance` — tab laporan pajak/penjualan/penagihan/arus kas/rincian
  piutang/tutup buku + ekspor.
- Sidebar & terjemahan (id/en) diperbarui.

## 9. Verifikasi (bukti)

- `go build ./...` — OK
- `go vet ./...` — OK
- `TESTCONTAINERS_RYUK_DISABLED=true go test ./... -count=1` — **seluruh paket OK**
- `TestPostgresMigrationsSmoke` — OK (termasuk tabel/kolom/sequence baru)
- Test repositori baru (`TestBillingOpsRepositories`,
  `TestInvoiceFilteredRepository`) — OK
- `bash scripts/check-connect-errors.sh` — pass
- `bash scripts/check-layer-boundaries.sh` — pass
- `buf lint` + `buf generate` — idempoten (checksum `api/gen` & `web/src/gen`
  tidak berubah antar-generate)
- `npx tsc -p tsconfig.app.json --noEmit` — bersih
- `npx vite build` — sukses
- ESLint file tersentuh — bersih

## 10. Sisa / Tindak Lanjut Opsional

1. UI khusus daftar transaksi gateway & settlement (hook/RPC sudah siap).
2. UI batch mass payment dengan pemilihan invoice di tabel (backend siap,
   tombol saat ini tersedia lewat API; dapat ditambahkan checkbox selection).
3. ConnectRPC untuk `collection.Service` (layanan usecase sudah ada untuk
   pengingat/isolir massal; saat ini dipakai internal/siap diekspos).
4. Adapter agregator Moneta masih feature-flagged menunggu dokumentasi API.
5. Roadmap Track B tetap di luar cakupan: collector/agent lapangan,
   inventori ONT, HR/absensi, FCM/APK/Telegram, GIS/ODP.
