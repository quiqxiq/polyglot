# Implementation Plan: Penyelarasan ISP Core (Customer, Subscription, Plan, Registrasi, Provisi MikroTik, & Sub-Message Protobuf Bertipe)

Rencana kerja komprehensif ini menyajikan pemisahan arsitektur bertipe kuat (*strongly-typed*) antara layanan **PPPoE** dan **Hotspot**, pendefinisian sub-message Protobuf bertipe (Opsi 2) untuk `Plan` dan `Subscription`, pengaturan pola penamaan kredensial dinamis, alur kerja registrasi & aktivasi lapangan oleh teknisi, serta pemetaan 1:1 ke RouterOS MikroTik.

---

## User Review Required

> [!IMPORTANT]
> **Poin Utama Desain Protobuf (Opsi 2: Typed Sub-Messages & Zero Waterfall)**:
> 1. **`Plan` Message**: Memisahkan field umum (bandwidth, harga, tax) dengan sub-message spesifik:
>    - `PPPoEPlanConfig` (`remote_address_pool`, `address_list`)
>    - `HotspotPlanConfig` (`ip_pool_name`, `shared_users`, `validity`, `validity_mode`, `expire_mode`, `lock_user`, `lock_server`, `limit_uptime`, `limit_bytes`, `selling_price`)
> 2. **`Subscription` Message**:
>    - Status operasional jaringan murni: `ACTIVE`, `ISOLATED`, `SUSPENDED`, `TERMINATED`, `PENDING`.
>    - Sub-message spesifik: `PPPoESubscriptionConfig` (`local_address`, `remote_address`, `rate_limit`) & `HotspotSubscriptionConfig` (`server`).
>    - Field denormalisasi tampilan untuk frontend: `customer_name`, `customer_phone`, `customer_code`, `plan_name`, `plan_price`, `device_name`, `device_host` (mencegah *N+1 waterfall query*).
> 3. **Alur Pemasangan & Pemilihan Router oleh Teknisi (`registration.proto`)**:
>    - Pendaftaran mandiri $\rightarrow$ status `PENDING` (pilih paket, tanpa memilih router/kredensial).
>    - Admin setujui & jadwalkan $\rightarrow$ status `APPROVED` (penugasan teknisi & tanggal pasang).
>    - Teknisi selesai pasang di rumah pelanggan $\rightarrow$ memilih Router BRAS target (`device_id`) $\rightarrow$ status `INSTALLED` / `ACTIVE` (memicu auto-provisi akun MikroTik + penerbitan invoice pertama + notifikasi WhatsApp).
> 4. **Pola Penamaan Username Otomatis**:
>    - Disimpan di `system_settings` (`isp.pppoe_username_pattern = '{initials}{digits4}'`, dll.) untuk menghasilkan kredensial unik (contoh: *Budi Santoso* $\rightarrow$ `bs4829`).

---

## 1. Desain Protobuf v1 Terstruktur (Opsi 2)

### A. `api/proto/v1/billing.proto`
```protobuf
// ── Konfigurasi Spesifik Paket PPPoE ──
message PPPoEPlanConfig {
  string remote_address_pool = 1; // Pool IP sumber dial pelanggan
  string address_list = 2;        // Address list firewall
}

// ── Konfigurasi Spesifik Paket Hotspot ──
message HotspotPlanConfig {
  string ip_pool_name = 1;        // IP Pool Hotspot Profile
  int32 shared_users = 2;         // Batas login bersamaan (default 1)
  string validity = 3;            // Masa aktif (misal "30d")
  string validity_mode = 4;       // "CALENDAR" | "UPTIME"
  string expire_mode = 5;         // "ntf" | "ntfc" | "rem" | "remc" | "0"
  bool lock_user = 6;             // Kunci MAC
  bool lock_server = 7;           // Kunci Server Hotspot
  string limit_uptime = 8;        // Kuota waktu (kosong = unlimited)
  string limit_bytes = 9;         // Kuota data (kosong = unlimited)
  double selling_price = 10;      // Harga jual voucher
}

// ── Master Paket Layanan ──
message Plan {
  string id = 1;
  string name = 2;                // Nama paket & profil MikroTik
  string service_type = 3;        // "PPPOE" | "HOTSPOT" | "DEDICATED"
  int32 bandwidth_download_kbps = 4;
  int32 bandwidth_upload_kbps = 5;
  int32 burst_download_kbps = 6;
  int32 burst_upload_kbps = 7;
  int32 burst_threshold_kbps = 8;
  int32 burst_time_seconds = 9;
  double price = 10;              // Harga dasar tagihan bulanan
  double installation_fee = 11;   // Biaya pasang
  double tax_percent = 12;
  string parent_queue = 13;       // Parent queue Queue Tree
  bool is_active = 14;
  string description = 15;

  // Sub-message bertipe
  PPPoEPlanConfig pppoe_config = 16;
  HotspotPlanConfig hotspot_config = 17;
}

// ── Konfigurasi Spesifik Langganan PPPoE ──
message PPPoESubscriptionConfig {
  string local_address = 1;
  string remote_address = 2;
  string rate_limit = 3;
}

// ── Konfigurasi Spesifik Langganan Hotspot ──
message HotspotSubscriptionConfig {
  string server = 1; // "all" atau nama server
}

// ── Langganan Pelanggan Aktif ──
message Subscription {
  string id = 1;
  string tenant_id = 2;
  string customer_id = 3;
  string plan_id = 4;
  string device_id = 5;           // UUID Router BRAS target
  string service_type = 6;        // "PPPOE" | "HOTSPOT" | "DEDICATED"
  string remote_username = 7;
  string router_profile = 8;      // Profil paket aktif di router (basis restore)
  
  string status = 9;              // "ACTIVE" | "ISOLATED" | "SUSPENDED" | "TERMINATED" | "PENDING"
  string provision_status = 10;   // "OK" | "PENDING" | "FAILED" | "NONE"
  
  string billing_cycle = 11;      // "MONTHLY"
  int32 billing_day = 12;         // Tanggal tagihan terbit (1-28)
  bool auto_isolate = 13;
  int32 isolation_grace_days = 14;
  double custom_price = 15;
  int64 start_date_unix = 16;
  int64 end_date_unix = 17;
  string notes = 18;

  // Sub-message spesifik
  PPPoESubscriptionConfig pppoe_config = 19;
  HotspotSubscriptionConfig hotspot_config = 20;

  // Denormalisasi Tampilan (Zero Waterfall di Frontend)
  string customer_name = 21;
  string customer_phone = 22;
  string customer_code = 23;
  string plan_name = 24;
  double plan_price = 25;
  string device_name = 26;
  string device_host = 27;
}
```

---

### B. `api/proto/v1/registration.proto`
```protobuf
message Registration {
  string id = 1;
  string registration_no = 2;   // REG-202608-0001
  string plan_id = 3;
  string plan_name = 4;
  double plan_price = 5;
  string full_name = 6;
  string phone = 7;
  string email = 8;
  string address = 9;
  double latitude = 10;
  double longitude = 11;
  string notes = 12;
  
  // Status: PENDING | APPROVED | INSTALLED | ACTIVE | REJECTED | CANCELLED
  string status = 13;

  int64 scheduled_install_date_unix = 14;
  string scheduled_install_time = 15;
  string assigned_technician_id = 16;
  string technician_name = 17;
  
  string target_device_id = 18;    // Router yang dipilih teknisi saat pasang
  string target_device_name = 19;
  int64 installed_at_unix = 20;
  string technician_notes = 21;

  string customer_id = 22;
  string subscription_id = 23;
  string invoice_id = 24;
}

message ConvertRegistrationRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  string device_id = 2 [(buf.validate.field).string.min_len = 1]; // Teknisi pilih router
  string technician_notes = 3;
}
```

---

## 2. Pemetaan Teknis Database vs MikroTik RouterOS

```
┌─────────────────────────────────────────────────────────────┐
│                    ServicePlan (Database)                   │
├──────────────────────────────┬──────────────────────────────┤
│ PPPoE Fields:                │ Hotspot Fields:              │
│ - name                       │ - name                       │
│ - RateLimitWithBurst()       │ - RateLimitWithBurst()       │
│ - remote_address_pool        │ - ip_pool_name               │
│ - parent_queue               │ - shared_users, validity     │
│ - address_list               │ - expire_mode, lock_user     │
└──────────────┬───────────────┴──────────────┬───────────────┘
               │ (Auto-Create)                │ (Auto-Create)
               ▼                              ▼
┌──────────────────────────────┐┌──────────────────────────────┐
│ MikroTik /ppp/profile        ││ MikroTik /ip/hotspot/profile │
└──────────────▲───────────────┘└──────────────▲───────────────┘
               │                              │
               │ (Profile Name)               │ (Profile Name)
┌──────────────┴───────────────┐┌──────────────┴───────────────┐
│ MikroTik /ppp/secret         ││ MikroTik /ip/hotspot/user    │
├──────────────────────────────┤├──────────────────────────────┤
│ name = remote_username       ││ name = remote_username       │
│ password = decrypted_secret  ││ password = decrypted_secret  │
│ local-address, remote-address││ server = "all" / hotspot_srv │
│ comment = polyglot:SUB-xxx   ││ comment = polyglot:SUB-xxx   │
└──────────────────────────────┘└──────────────────────────────┘
```

---

## 3. Proposed Changes Summary

### 1. Database & Migrasi SQL
#### [NEW] `migrations/000023_add_username_pattern_settings.up.sql` & `.down.sql`
- Menambahkan setting default: `isp.pppoe_username_pattern = '{initials}{digits4}'`, `isp.pppoe_username_prefix = ''`, `isp.pppoe_password_mode = 'digits6'`.

### 2. Protobuf Layer
#### [MODIFY] [billing.proto](file:///home/quixiq/projects/polyground/polyglot/api/proto/v1/billing.proto)
- Menerapkan Opsi 2 (`PPPoEPlanConfig`, `HotspotPlanConfig`, `PPPoESubscriptionConfig`, `HotspotSubscriptionConfig`, dan denormalized display fields).
#### [MODIFY] [registration.proto](file:///home/quixiq/projects/polyground/polyglot/api/proto/v1/registration.proto)
- Memperbarui `ConvertRegistrationRequest` dengan kewajiban `device_id`.
#### [MODIFY] [customer.proto](file:///home/quixiq/projects/polyground/polyglot/api/proto/v1/customer.proto)
- Menyederhanakan status customer ke identitas member (`ACTIVE`).

### 3. Domain & Port Layer
#### [NEW] `internal/domain/subscription/provision_pppoe.go`
- Struct `PPPoEProvisionSpec`, `PPPoESecretSpec`, `PPPoEProfileSpec`.
#### [NEW] `internal/domain/subscription/provision_hotspot.go`
- Struct `HotspotProvisionSpec`, `HotspotUserSpec`, `HotspotProfileSpec`.
#### [MODIFY] [router_account_manager.go](file:///home/quixiq/projects/polyground/polyglot/internal/port/router_account_manager.go)
- Interface method spesifik: `ProvisionPPPoE`, `ProvisionHotspot`, `ProvisionDedicated`.

### 4. Usecase & Mapper Layer
#### [MODIFY] [convert.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/registration/convert.go) & [convert_artifacts.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/registration/convert_artifacts.go)
- Username generator dinamis dari inisial + digits.
- Provisi MikroTik langsung terpanggil saat teknisi menyelesaikan pemasangan dengan `device_id`.
#### [MODIFY] [plan_account.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/billing/plan_account.go)
- Mappers terisolasi untuk `PPPoEProvisionSpec` dan `HotspotProvisionSpec`.
#### [MODIFY] [billing_handler.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/connect/billing/billing_handler.go)
- Mapping GORM model <-> Protobuf sub-message Opsi 2.

### 5. Frontend React SPA
#### [MODIFY] [web/src/features/](file:///home/quixiq/projects/polyground/polyglot/web/src/features/)
- Form Plan adaptif: menampilkan input PPPoE vs Hotspot sesuai `service_type`.
- Modal Penyelesaian Teknisi: dropdown pemilihan Router BRAS target (`device_id`).

---

## Verification Plan

### Automated Tests
1. **Proto Lint & Generation**:
   ```bash
   make proto-check
   ```
2. **Boundary & Error Checks**:
   ```bash
   make check-connect-errors check-layer-boundaries
   ```
3. **Backend Unit & Integration Tests**:
   ```bash
   go test ./internal/domain/... ./internal/usecase/... ./internal/adapter/... -v -count=1
   ```
4. **Frontend Build**:
   ```bash
   pnpm --dir web build
   ```
