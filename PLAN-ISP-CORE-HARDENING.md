# PLAN-ISP-CORE-HARDENING — Perbaikan Menyeluruh Inti ISP Platform

> Status: Aktif
> Basis: hasil analisis inti project (billing, subscription, payment gateway, cashbook, provisioning MikroTik, isolir, scheduler, portal pelanggan)
> Cakupan: seluruh temuan (kritis, tinggi, sedang, rendah) — dikelompokkan per fase, milestone bertahap.
> Keputusan desain: payment gateway diarahkan **multi-gateway** (Tripay + Midtrans + Xendit) melalui port `PaymentGateway` yang sudah ada.

## Konvensi

- Setiap task punya ID unik (`F<fase>-<nomor>`) untuk tracking.
- Gate setiap fase: `make build && make vet && make test && make lint && make check-connect-errors check-layer-boundaries`.
- Perubahan skema DB wajib melalui migrasi baru (jangan edit migrasi lama yang sudah dirilis).
- Perubahan protobuf wajib `buf lint` + `make proto-check`.
- Fase 0 adalah prasyarat seluruh fase; Fase 1 adalah blocker produksi.

---

## Fase 0: Fondasi & Infrastruktur Test (Prasyarat)

Tujuan: membuat test harness yang belum ada, agar setiap perbaikan diverifikasi tanpa regresi. Test yang mengunci perilaku benar tetapi masih bug ditandai `t.Skip("RED TEST — …")` beserta ID task perbaikannya; gate `make test` tetap hijau sampai fase terkait mendarat, lalu `t.Skip` dihapus satu per satu.

| ID | Task | File Target | Dependensi | Verifikasi |
|---|---|---|---|---|
| F0-1 | Buat unit test suite `internal/adapter/tripay/` — saat ini nol test. Mock HTTP transport (`httptest`), uji `CreateCharge`, `ParseWebhook`, `CheckStatus`: signature, ExternalID mapping, nominal, error path | `internal/adapter/tripay/adapter_test.go` | - | `go test ./internal/adapter/tripay/ -v` |
| F0-2 | Tambah test case di `internal/usecase/billing/gateway_test.go`: callback dengan ExternalID berbeda dari create, pembayaran parsial tidak memicu restore, invoice PAID → reject | `internal/usecase/billing/gateway_test.go` | F0-1 | `go test ./internal/usecase/billing/ -run TestGateway -v` |
| F0-3 | Buat test `internal/adapter/provisioner/`: siklus isolir→restore (pastikan `disabled=no` setelah restore), suspend→resume PPP & hotspot, konsistensi penamaan queue dedicated | `internal/adapter/provisioner/provisioner_test.go` (extend) | - | `go test ./internal/adapter/provisioner/ -v` |
| F0-4 | Buat test `internal/usecase/registration/convert_test.go`: happy path + partial failure (error DB di langkah 3/4 harus rollback penuh) | `internal/usecase/registration/convert_test.go` | - | `go test ./internal/usecase/registration/ -v` |
| F0-5 | Test portal OTP: `ConsumeOTP` benar-benar terkunci setelah max attempts; `consumed_at` persist saat OTP cocok | `internal/adapter/postgres/portal_repository_test.go` | - | `go test ./internal/adapter/postgres/ -run TestPortalOTP -v` |

Estimasi: 2–3 hari. Pattern referensi: `internal/usecase/billing/testutil_test.go`, `internal/adapter/provisioner/provisioner_test.go`.

---

## Fase 1: Blocker Produksi — Payment Gateway & Provisioning Router (KRITIS)

> **Status: SELESAI** — seluruh task F1-1..F1-13 diimplementasikan; test Fase 1 aktif dan hijau (full suite `-race -cover`, lint + boundary checks bersih).

Catatan implementasi:
- **F1-4** tuntas sebagai konsekuensi F1-1: `gateway_transactions.external_id` kini menyimpan Tripay reference, dan `CheckStatus`/callback memakai kolom itu.
- **F1-5** memakai `Save(tx)` penuh (bukan `UpdateStatus`): `raw_callback`, `callback_count++`, `paid_at`, `payment_channel`, `expires_at` ikut terpersistensi.
- **F1-6** diperluas: selain driver `NewSetUserCommand` menulis `disabled=yes/no`, jalur restore/resume (`UpdateAccount`) mengeset `disabled=false` eksplisit.
- **F1-7** tuntas: `UpdateAccount` PPP selalu mengirim `SetSecretDisabled(false)` saat pindah profil.
- **F1-9** tuntas via pencocokan profil case-insensitive + `ensurePlanProfile` mengembalikan nama profil aktual di router; fallback setting diubah ke `isolir` agar selaras seed migrasi.
- **F1-10/F1-11** tuntas: `isolirAccount` membawa `address-list`; profil isolir di-throttle `64k/64k` (bukan `0/0`).
- **F1-12** tuntas dengan port baru `port.ConversionWriter` + `postgres.ConversionWriter` (satu `gorm.Transaction`) dan fake `mocktest.FakeConversionWriter`; provisi router dipindah pasca-commit (best-effort).
- **Bonus**: pemetaan error webhook HTTP diperbaiki — 401 hanya untuk signature invalid, 404 untuk transaksi tak dikenal, selainnya 500 (sebelumnya semua error dibalas 401).

### Sprint 1A — Perbaikan Tripay

| ID | Task | Detail | File |
|---|---|---|---|
| F1-1 | Fix ExternalID mapping: simpan `data.reference` (ID Tripay) sebagai `ExternalID` saat CreateCharge; tambah `MerchantRef` di `ChargeResult` | `internal/adapter/tripay/adapter.go:162-170` |
| F1-2 | Fix signature callback sesuai dokumentasi resmi Tripay: `HMAC_SHA256(private_key, raw_body)` | `internal/adapter/tripay/adapter.go:187-198` |
| F1-3 | Fix nominal pelunasan: gunakan `total_amount`, bukan `amount_received`; samakan dengan jalur `CheckStatus` | `internal/adapter/tripay/adapter.go:208-211` |
| F1-4 | Fix `CheckStatus` reference: kirim Tripay reference dari `gateway_transactions.external_id` | `internal/adapter/tripay/adapter.go:227` |
| F1-5 | Persist metadata callback: `raw_callback`, `callback_count++`, `paid_at`, `payment_channel`, `expires_at` | `internal/adapter/postgres/gateway_transaction_repository.go:60-70`; `internal/usecase/billing/gateway.go:128-135` |

### Sprint 1B — Provisioning Router (Suspend/Resume/Isolir)

| ID | Task | Detail | File |
|---|---|---|---|
| F1-6 | Fix hotspot `NewSetUserCommand`: tangani argumen `disabled` (yes/no); pastikan suspend→restore/resume menulis `disabled=no` kembali | `internal/driver/mikrotik/hotspot/commands.go:112-123` |
| F1-7 | Fix PPP resume/restore: panggil `SetSecretDisabled(..., false)` pada restore/resume | `internal/adapter/provisioner/pppoe.go`; `internal/adapter/provisioner/isolation.go:124-138` |
| F1-8 | Guard OnPaid: restore hanya bila invoice `PAID` | `internal/adapter/provisioner/isolation.go:204-211` |
| F1-9 | Satukan case profil isolir (seed `isolir` vs fallback `ISOLIR`) | `internal/port/setting_reader.go:26-27`; migrasi baru |
| F1-10 | Tambah `address-list` ke profil isolir + `ensurePlanProfile` update profil existing | `internal/adapter/provisioner/plan_sync.go:86-88,112-131` |
| F1-11 | Ganti rate isolir `0/0` (unlimited) menjadi rate terbatas | `internal/adapter/provisioner/plan_sync.go:86-88` |

### Sprint 1C — Konversi Registrasi Atomik

| ID | Task | Detail | File |
|---|---|---|---|
| F1-12 | Bungkus `ConvertWithDevice` dalam satu DB transaction (customer+subscription+invoice+registration); provisi router tetap best-effort pasca-commit | `internal/usecase/registration/convert.go:65-107` |
| F1-13 | Biaya instalasi masuk `Subtotal`/`TaxAmount`/`Total` invoice pertama | `internal/usecase/registration/convert_artifacts.go:168-218` |

Estimasi: 5–7 hari.

---

## Fase 2: Integritas Provisioning & Lifecycle (TINGGI)

> **Status: SELESAI** — F2-1..F2-13 tuntas; test terkait aktif dan hijau.

Catatan implementasi:
- **F2-1/F2-5/F2-9/F2-10**: nama queue dedicated konsisten `dq-<username>`; `ChangePlan` memakai `RateLimitWithBurst()`; `caller-id` (MAC) PPP dan `mac-address`/`address` hotspot kini ditulis ke command builder.
- **F2-2/F2-3**: guard transisi — `Activate` hanya dari PENDING, `ChangePlan` hanya ACTIVE/ISOLATED + reset override `rate_limit`. `ChangePlan` pada langganan ISOLATED hanya menyiapkan profil baru tanpa memindahkan akun keluar dari profil isolir.
- **F2-4**: `ensurePlanProfile` membandingkan rate/burst, parent-queue, address-list, dan pool lalu meng-update profil existing; field kosong tidak menimpa profil dashboard.
- **F2-6**: setting `isp.pppoe_username_pattern` / `prefix` / `password_mode` dibaca oleh `ManageSubscription.Create` dan `ConvertUseCase`; `idgen.Password` mendukung `digits6` (default), `digits4`, `random8`, `phone`.
- **F2-7**: kolom baru `registrations.target_device_id` (migrasi 000025); `MarkInstalled` menyimpannya dan konversi memakainya sebagai fallback router.
- **F2-8/F2-12**: provisi registrasi memakai `planUC.Build*ProvisionSpec` (burst, timeout, pool, config bertipe); DEDICATED tidak lagi dipetakan menjadi PPPOE.
- **F2-11**: `auto_isolate` dan `isolation_grace_days` per-langganan dihormati; semua jalur pembuatan langganan men-set `AutoIsolate=true` secara default.
- **F2-13**: kegagalan restore pasca-bayar menandai `provision_status=FAILED`; lifecycle worker me-restore ulang dan memulihkan status ACTIVE pada siklus berikutnya.

| ID | Task | Detail | File |
|---|---|---|---|
| F2-1 | Konsistensi nama queue dedicated: `dq-<username>` di `BuildDedicatedProvisionSpec` | `internal/usecase/plan/plan_account.go:154` |
| F2-2 | Guard transisi `Activate` (hanya dari PENDING) | `internal/usecase/subscription/lifecycle.go:49-59` |
| F2-3 | Guard transisi `ChangePlan` (ACTIVE/ISOLATED) + reset `RateLimit` override | `internal/usecase/subscription/lifecycle.go:97-108` |
| F2-4 | `ensurePlanProfile` update profil existing (rate/burst/address-list/parent) | `internal/adapter/provisioner/plan_sync.go:112-131` |
| F2-5 | `ChangePlan` pakai `RateLimitWithBurst()` | `internal/usecase/subscription/lifecycle.go:115` |
| F2-6 | Hubungkan setting `isp.pppoe_username_pattern/prefix/password_mode` ke generator kredensial | `manage_subscription.go:200-205`; `convert_artifacts.go:73` |
| F2-7 | Simpan `target_device_id` teknisi saat `MarkInstalled` (kolom DB + domain + handler + mapper) | `registration_handler.go:127-138`; migrasi baru |
| F2-8 | Registrasi memakai `planUC.Build*ProvisionSpec` (bukan spec manual) | `convert_artifacts.go:110-160` |
| F2-9 | Tulis `caller-id` (MAC) pada PPP secret command builder | `ppp/commands.go:11-56`; `port/ppp_session.go:17-27` |
| F2-10 | Tulis `mac-address`/`address` pada hotspot user command | `hotspot/commands.go:95-123`; `hotspot/types.go:29-39` |
| F2-11 | Hormati `auto_isolate`/`isolation_grace_days` per-subscription di worker | `internal/usecase/billing/isolate_worker.go:141-149` |
| F2-12 | DEDICATED tidak kehilangan tipe saat konversi registrasi | `convert_artifacts.go:28-35,110-160` |
| F2-13 | Retry restore pasca-bayar: set `provision_status=FAILED` bila restore gagal | `internal/adapter/provisioner/isolation.go:227-238` |

Estimasi: 5–7 hari.

---

## Fase 3: Billing & Notifikasi (TINGGI–SEDANG)

> **Status: SELESAI** — F3-1..F3-10 tuntas; test terkait hijau. Hanya test RED F5-1 (lockout OTP) yang masih menunggu Fase 5.

Catatan implementasi:
- **F3-1**: `ReminderWorker` + cron baru `REMINDER_CRON` (default `0 7 * * *`) — reminder H-7/H-3/H-1/hari-H, idempoten per invoice per hari via `ExistsForInvoiceSince`.
- **F3-2**: tagihan `UNPAID` lewat jatuh tempo ditandai `OVERDUE`; isolir menerima `UNPAID`/`OVERDUE`/`PARTIAL`.
- **F3-3**: invoice hanya diterbitkan pada/setelah `billing_day` (di-clamp ke akhir bulan pendek) untuk periode berjalan; backfill periode lampau tidak digate.
- **F3-4**: `PAYMENT_RECEIPT` dirender dari `notification_templates` + fallback teks bawaan.
- **F3-5**: `BillingService/CancelInvoice` terdaftar di RBAC (`billing:manage`).
- **F3-6**: port `InvoiceCanceller` + `postgres.InvoiceCanceller` — pembatalan + jurnal kas koreksi (OUT) atomik; seed kategori `cc-refund` (migrasi 000027).
- **F3-7**: `Update` menandai `provision_status=FAILED` bila sinkronisasi router gagal (retry worker).
- **F3-8**: soft delete customer/subscription/invoice + filter `deleted_at IS NULL` di seluruh query baca.
- **F3-9/F3-10**: migrasi 000026 — unique `(subscription_id, period)` dan unique `(device_id, remote_username) WHERE deleted_at IS NULL`; diuji di smoke test PostgreSQL.

| ID | Task | Detail | File |
|---|---|---|---|
| F3-1 | Implementasi `BILL_REMINDER` (H-7/H-3/H-1) via worker + scheduler | baru: `internal/usecase/billing/reminder_worker.go` |
| F3-2 | Implementasi status `OVERDUE` + `overdueInvoice` terima `OVERDUE`/`PARTIAL` | `isolate_worker.go:214` |
| F3-3 | `billing_day` menentukan kapan invoice diterbitkan | `run_billing.go:57-67` |
| F3-4 | `PAYMENT_RECEIPT` memakai template DB, bukan hardcode | `payment_processor.go:129-139,180-183` |
| F3-5 | Daftarkan `CancelInvoice` di RBAC | `procedure_permissions.go:231-236` |
| F3-6 | Cancel invoice PARTIAL membuat jurnal koreksi kas OUT | `manage_invoice.go:67-81` |
| F3-7 | `manage_subscription.Update` set `provision_status=FAILED` bila re-provision gagal | `manage_subscription.go:352-376` |
| F3-8 | Implementasi soft delete konsisten (set `deleted_at` + filter) atau hapus kolom menganggur | repository terkait |
| F3-9 | Unique `(subscription_id, period)` di `invoices` | migrasi baru |
| F3-10 | Unique `(device_id, remote_username) WHERE deleted_at IS NULL` di `subscriptions` | migrasi baru |

Estimasi: 4–6 hari.

---

## Fase 4: Multi-Gateway Payment (SEDANG)

> **Status: SELESAI** — F4-1..F4-7 tuntas (backend + UI settings gateway); test adapter, registry, dan guard hijau.

Catatan implementasi:
- **F4-1**: `GatewayRegistry` (`internal/usecase/billing/gateway_registry.go`) — default dari setting `gw.active`, fallback gateway enabled pertama; usecase memakai registry via `NewGatewayChargeUseCaseWithRegistry` (konstruktor lama tetap dipertahankan untuk kompatibilitas).
- **F4-2**: webhook per gateway `/api/webhook/{tripay|midtrans|xendit}`; header verifikasi dipilih per gateway (Xendit `x-callback-token`, Tripay `X-Callback-Signature`, Midtrans signature di body); endpoint charge/check-charge menerima `gateway` opsional dan mengembalikan nama gateway; pemetaan error 401/404.
- **F4-3**: adapter Midtrans — Snap create, Core API check-status, verifikasi signature SHA-512 `order_id + status_code + gross_amount + server_key`; `ExternalID` = `order_id`.
- **F4-4**: adapter Xendit — Invoice API create/check-status, verifikasi `x-callback-token` constant-time; `ExternalID` = id invoice, `MerchantRef` = `external_id` (lookup webhook fallback ke MerchantRef).
- **F4-5**: migrasi 000028 seed `gw.active`, `gw.midtrans.*`, `gw.xendit.*` (kategori `isp_gateway`) + halaman **Settings → Payment Gateway** (generik per prefix, input password untuk key sensitif, switch untuk `*.enabled`).
- **F4-6**: `reusePendingCharge` — transaksi PENDING untuk invoice+gateway yang sama dipakai ulang (idempoten); PENDING gateway lain ditandai `EXPIRED` agar tidak ada dua tagihan aktif.
- **F4-7**: `SettingRepository.WithVault` — key sensitif (`*.api_key`, `*.private_key`, `*.secret_key`, `*.server_key`, `*.callback_token`) dienkripsi AES-GCM dengan prefix `enc:v1:` dan didekripsi transparan saat dibaca; tanpa vault perilaku lama tetap berjalan.

| ID | Task | Detail | File |
|---|---|---|---|
| F4-1 | `GatewayRegistry` (map name→gateway); usecase charge memakai registry | baru: `internal/usecase/billing/gateway_registry.go` |
| F4-2 | Webhook routing per path (`/api/webhook/tripay|midtrans|xendit`) | `internal/adapter/http/gateway/handler.go` |
| F4-3 | Adapter Midtrans (Snap/Core API) | baru: `internal/adapter/midtrans/adapter.go` |
| F4-4 | Adapter Xendit (Invoice API) | baru: `internal/adapter/xendit/adapter.go` |
| F4-5 | Setting `gw.midtrans.*`, `gw.xendit.*` + UI admin | migrasi baru; setting UI |
| F4-6 | Guard duplikat transaksi gateway PENDING per invoice | `internal/usecase/billing/gateway.go:43-90` |
| F4-7 | Enkripsi kredensial gateway (AES-GCM vault) | vault; migrasi |

Estimasi: 5–8 hari.

---

## Fase 5: Portal Pelanggan, Keamanan, & UX (SEDANG)

> **Status: SELESAI** — F5-1..F5-11 tuntas; seluruh test RED (termasuk OTP lockout) aktif dan hijau.

Catatan implementasi:
- **F5-1**: `ConsumeOTP` meng-commit `attempts++`/`consumed_at` lebih dulu, lalu mengembalikan `ErrOTPLocked`/`ErrOTPNotFound` setelah transaksi sukses — lockout benar-benar tersimpan.
- **F5-2**: halaman `/portal/login` (OTP WA) dan `/portal` (dashboard: status layanan, langganan, tagihan menunggak + bayar online, riwayat pembayaran); halaman isolir PPPoE kini menyematkan langkah verifikasi OTP sebelum membayar.
- **F5-3**: `Login` mengembalikan `expiresAt` sesi aktual (setting `isp.portal_session_hours`) dan dipakai response ConnectRPC + HTTP.
- **F5-4**: `/api/portal/charge` pindah dari publik ke endpoint bersesi (`Bearer` token portal) dengan verifikasi kepemilikan invoice (`InvoiceForCustomer`, sentinel `ErrPortalForbidden`); invoice orang lain/tidak ada sama-sama 403.
- **F5-5**: `PublicBillView` tanpa sesi tidak lagi mengembalikan `customer_code` (hanya nama tersamar + rincian tagihan).
- **F5-6**: `ResolveByPortalCode` menerima `UNPAID`/`OVERDUE`/`PARTIAL` (dengan outstanding > 0).
- **F5-7**: OTP portal masuk antrean `wa_notifications` (`PORTAL_OTP`) lewat worker retry; fallback kirim langsung hanya bila repo antrean tidak tersedia.
- **F5-8**: fallback kunci enkripsi hardcoded dihapus dari vault & model — key wajib 32 byte, dekripsi dengan key salah menjadi error eksplisit.
- **F5-9**: `DeviceModel` memakai kolom produksi `extra` (JSONB) & `tags` (TEXT[]) — diverifikasi round-trip di PostgreSQL nyata (`device_repository_pg_test.go`).
- **F5-10**: `AutoMigrate` di konstruktor `NewCustomerRepository`/`NewInvoiceRepository` dihapus (hanya lewat gate di `store.go`).
- **F5-11**: script integrasi RouterOS mem-persist `webhook_token` ke device dan menyertakan `?device=<id>`; handler webhook memverifikasi token constant-time (fallback legacy `rtr_<id>` untuk skrip lama).

| ID | Task | Detail | File |
|---|---|---|---|
| F5-1 | Fix bug rollback lockout OTP (`ConsumeOTP`) | `portal_repository.go:77-125` |
| F5-2 | Halaman login/dashboard portal (route + hook yang sudah ada) | `web/src/routes/portal/` |
| F5-3 | `expires_at_unix` portal ambil dari setting `portal_session_hours` | `portal_handler.go:46` |
| F5-4 | Auth `/api/portal/charge` + verifikasi kepemilikan invoice | `http/gateway/handler.go:25-28` |
| F5-5 | Batasi informasi `/api/portal/bill` tanpa sesi | `http/portal/handler.go:30` |
| F5-6 | `ResolveByPortalCode` terima `PARTIAL` | `checkout.go:57-62` |
| F5-7 | OTP via antrean WA (bukan kirim langsung) | `manage_portal.go:74-77` |
| F5-8 | Hapus fallback kunci enkripsi hardcoded | `device_repository.go:151-162`; `device_model.go` |
| F5-9 | Fix kolom `extra_json`/`tags_json` → `extra`/`tags` | `model/device_model.go:27-28` |
| F5-10 | Hapus AutoMigrate bocor di konstruktor repository | `customer_repository.go:24`; `invoice_repository.go:22` |
| F5-11 | Verifikasi token webhook RouterOS | `http/webhook/handler.go:53` |

Estimasi: 4–6 hari.

---

## Fase 6: Hardening, Observabilitas, & Konsistensi (RENDAH–SEDANG)

> **Status: SELESAI** — F6-1..F6-14 tuntas; seluruh gate hijau (`make build`, `make vet`, `go test ./... -race -cover`, `make lint`).

Catatan implementasi:
- **F6-1**: `port.JobLocker` + `postgres.AdvisoryLocker` (`pg_try_advisory_lock` pada koneksi tertahan; no-op di dialect non-Postgres); setiap job cron dibungkus lock — siklus dilewati bila instance lain memegang lock; diuji mutual-exclusion di PostgreSQL nyata.
- **F6-2**: migrasi `000030_document_number_sequences` + `nextDocumentSeq` — nomor `PAY-`/`TRX-` kini dari sequence Postgres (fallback `idgen.Digits` di sqlite), bebas tabrakan pada pembayaran/cancel paralel.
- **F6-3**: `CleanupIsolationAddressList` dipanggil best-effort saat `Terminate` (lifecycle), delete subscription, dan delete customer.
- **F6-4**: worker isolasi melanjutkan siklus saat satu subscription gagal (`IsolationResult.Errors`, log + `continue`).
- **F6-5**: `MarkInstalled`/`Reject`/`Cancel` registrasi mengantre notifikasi WA (`INSTALLATION_COMPLETED`, `REGISTRATION_REJECTED`, `REGISTRATION_CANCELLED`) lewat `queueTemplate` dengan fallback in-code; template di-seed migrasi `000029`.
- **F6-6**: `DeleteIsolationInfrastructure` menghapus profil isolir PPP/hotspot + opsional redirect/filter/walled-garden; `DeleteIsolationProfile` usecase aktif.
- **F6-7**: `Isolate` otomatis memasang `EnsureIsolationFilter` saat redirect + address-list tersedia.
- **F6-8**: `port.PageFilter` + `FindPaged` pada repository customer/subscription/invoice; usecase list memakai tenant `tenant-default`.
- **F6-9**: perhitungan data (billing run, worker isolasi, reminder) memakai UTC; cron tetap mengikuti TZ server untuk jam operasional.
- **F6-10**: `ServicePlanModel` memetakan seluruh field domain (selling price, validity, expire/lock, limit).
- **F6-11**: `RecomputeDaily` memakai predikat range (`>= ? AND < ?`) agar index-friendly.
- **F6-12**: migrasi `000031_status_check_constraints` menambah CHECK pada status subscription/invoice/payment + service_type; pelanggaran diuji di smoke test.
- **F6-13**: `.env.example` mendokumentasikan seluruh spec cron.
- **F6-14**: dead code dihapus (`ppp.IsolirProfileParams`, `InvoiceUseCase.CreateInvoice/PayInvoice`, `tripay.Config.CallbackAction`); `MediaCleanerWorker` ter-wire di `app.go`.

| ID | Task | Detail | File |
|---|---|---|---|
| F6-1 | Leader election scheduler (advisory lock PostgreSQL) | `internal/app/scheduler.go:46-64` |
| F6-2 | Nomor `PAY-`/`TRX-` anti-tabrakan (sequence/atomic counter) | `payment_processor.go:90,108` |
| F6-3 | `Terminate` bersihkan address-list | `isolation.go:153-168` |
| F6-4 | Lifecycle worker toleran error per-subscription | `isolate_worker.go:78-83,171-173` |
| F6-5 | Notifikasi `MarkInstalled`/`Reject`/`Cancel` registrasi | `manage_registration.go:117-174` |
| F6-6 | Implementasi `DeleteIsolationProfile` | `manage_isolation.go:58-64` |
| F6-7 | Worker otomatis pasang `EnsureIsolationFilter` | `isolate_worker.go:162-164` |
| F6-8 | Paginasi + tenant filter repository | `*_repository.go` FindAll |
| F6-9 | Standarisasi zona waktu UTC (cron, due date, worker, snapshot) | `scheduler.go:41`; `run_billing.go:107` |
| F6-10 | Lengkapi mapping `ServicePlanModel` (selling price, validity, dll) | `model/service_plan_model.go:11-36` |
| F6-11 | Query laporan index-friendly (range, bukan `::date`) | `reporting_repository.go:77,87,96` |
| F6-12 | CHECK constraint status di DB | migrasi baru |
| F6-13 | Dokumentasi env cron di `.env.example` | `.env.example` |
| F6-14 | Hapus dead code (status, helper, deprecated) | berbagai file |

Estimasi: 5–7 hari.

---

## Timeline & Dependensi

```
Fase 0 ━━━━━━━━ (2-3 hari)  ← PRASYARAT semua fase
   ║
   ╠══► Fase 1A (3 hari) ── Tripay fix
   ║       ║
   ╠══► Fase 1B (3 hari) ── Router provisioning fix  ← paralel dengan 1A
   ║       ║
   ╠══► Fase 1C (1 hari) ── Registrasi atomik        ← paralel dengan 1A/1B
   ║
   ╠══► Fase 2 (5-7 hari) ── Lifecycle & integritas  ← setelah Fase 1
   ║
   ╠══► Fase 3 (4-6 hari) ── Billing & notifikasi    ← setelah Fase 1
   ║       (Fase 2 dan 3 bisa paralel)
   ║
   ╠══► Fase 4 (5-8 hari) ── Multi-gateway           ← setelah Fase 1A
   ║
   ╠══► Fase 5 (4-6 hari) ── Portal & keamanan       ← setelah Fase 1 & 3
   ║
   ╚══► Fase 6 (5-7 hari) ── Hardening               ← setelah semua
```

Total estimasi: ~30–45 hari kerja; ~15–20 hari kalender dengan 2–3 developer paralel.

## Urutan Eksekusi Disarankan

1. Minggu 1: Fase 0 + mulai Fase 1A/1B/1C paralel
2. Minggu 2: selesaikan Fase 1 + mulai Fase 2 & 3 paralel
3. Minggu 3: selesaikan Fase 2 & 3 + mulai Fase 4
4. Minggu 4: Fase 4 + Fase 5
5. Minggu 5: Fase 6 + integrasi end-to-end

## Gate Setiap Fase

```text
make build
make vet
make test
make lint
make check-connect-errors check-layer-boundaries
```

Integrasi perubahan jaringan: `make test-integration` (butuh PostgreSQL/Redis).
