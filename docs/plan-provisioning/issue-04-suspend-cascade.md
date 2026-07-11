# Issue 04: Suspend / Resume / Terminate Cascade

## Konteks

Temuan D di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` §5 menunjukkan satu fakta
penting dari repo referensi (`gembok-bill/config/serviceSuspension.js`): **satu
event bisnis "suspend pelanggan X" bisa memicu LEBIH DARI SATU target
provisioning sekaligus**, bukan cuma satu. Contoh nyata: PPPoE dipindah ke profil
`isolir` di MikroTik (isolir utama = `change_profile`, bukan sekadar `disable`
secret — K11), DAN untuk pelanggan static IP alamatnya dimasukkan ke address-list
isolir. Semua terjadi dalam satu transaksi bisnis.

**Catatan lapangan tentang GenieACS/TR-069:** repo referensi menyebut opsi
men-disable WAN di CPE lewat TR-069, tetapi di praktik produksi cascade suspend
**hampir selalu cukup dilakukan di lapisan MikroTik** — disable WAN via ACS
JARANG dipakai dan perilakunya tidak seragam antar-vendor ONU. Karena itu target
`genieacs_tr069` di issue ini bersifat **OPSIONAL/feature-flag** (per plan/vendor),
default hanya untuk reboot/monitoring, bukan pemutusan layanan. Bila fitur ini
diaktifkan, path native yang dipakai adalah `WANIPConnection.1.Enable` /
`WANPPPConnection.1.Enable=false` dan WAJIB diuji per vendor — pengetahuan
multi-path itu milik driver (K13), bukan usecase. Detail lihat §Ruang Lingkup &
Task 2.

Analisis itu mengonfirmasi desain `provisioning_sync_log` (`DATABASE-SCHEMA.md`
§6.3) sudah benar untuk kasus ini: satu `subscription_id` boleh punya BANYAK
baris sync untuk SATU event, masing-masing dengan `target_type` berbeda
(`mikrotik_ppp_secret`, `mikrotik_address_list`, `genieacs_tr069`) tapi ditulis
atomik dalam satu transaksi. Tidak ada perubahan skema — yang perlu dipastikan
adalah usecase Go benar-benar menulis multi-baris saat suspend/resume/terminate,
bukan satu baris saja. Issue ini menutup temuan #9 (multi-baris sync per event).

Issue ini melengkapi lifecycle subscription: Issue 03 menangani create → activate
(provision `/ppp secret`), sedangkan issue ini menangani transisi status
`active → suspended → active` dan `* → terminated`, beserta cascade
provisioning-nya dan audit trail lewat `subscription_status_history`
(`DATABASE-SCHEMA.md` §6.2).

## Prasyarat

- **Issue 01 (Provisioning Sync Engine)** — wajib. Sync Engine yang membaca baris
  `provisioning_sync_log` status `pending`, menerjemahkan jadi `command.Command`,
  memanggil `usecase/network.ExecuteCommand`, dan menautkan hasil ke
  `command_audit_log`. Issue ini hanya MENULIS baris `pending`; pemrosesan ke
  perangkat sepenuhnya milik Issue 01 (konvensi K4).
- **Issue 03 (Subscription Provisioning Lifecycle)** — wajib. Menyediakan
  provisioning `/ppp secret` PPPoE dan pola penulisan baris sync untuk
  `mikrotik_ppp_secret` yang dipakai ulang di sini.
- **Issue 08/09 (GenieACS integrasi penuh)** — opsional/menyusul. Kolom
  `genieacs_device_id` di `subscriptions` baru ditambahkan oleh Issue 08. Selama
  belum ada, baris `target_type=genieacs_tr069` cukup DITULIS `pending` dan akan
  diproses saat Issue 09 siap; kalau `genieacs_device_id` kosong, baris itu
  dilewati (tidak ditulis). Lihat §Ruang Lingkup.
- **Infra isolir per-router (one-time)** — wajib ada sebelum suspend pertama yang
  memakai address-list. Address-list isolir TIDAK memblokir apa pun tanpa firewall
  rule `src-address-list=LIST_ISOLIR action=drop` + NAT redirect portal + allow
  DNS/portal, dipasang idempoten (by comment) satu kali per router (K11). Pemasangan
  ini idealnya bagian setup device (Issue 02); issue ini menyediakan Task 9
  ("ensure isolir firewall") sebagai jaring pengaman bila belum terpasang.
- **Foundation Phase 4/5** — router Gin nyata, middleware `AuthRequired` +
  `RBACRequired`, `internal/adapter/http/dto/`, registry aktif.

## Ruang Lingkup

**In scope:**
- Tiga usecase orkestrasi di `internal/usecase/network/`: suspend, resume,
  terminate subscription.
- Penentuan target provisioning (cascade) berdasarkan data subscription
  (`pppoe_username`, `static_ip`, `genieacs_device_id`).
- Penulisan MULTI-BARIS `provisioning_sync_log` status `pending` dalam SATU
  transaksi bisnis (atomik) bersama update status subscription +
  `subscription_status_history`.
- Kebijakan "minimal target terpenuhi" untuk menentukan kapan status
  subscription boleh berubah (lihat Task 2).
- Konfigurasi kebijakan isolir MikroTik: default **`change_profile` ke profil
  ISOLIR** (bukan sekadar `disable` secret), sesuai K11 — profil isolir membawa
  redirect portal + rate-limit + hook address-list yang tidak didapat dari
  `disabled=yes` saja. `disable` disediakan sebagai mode alternatif, bukan default.
  Nama profil isolir & profil normal-saat-resume berasal dari plan/subscription
  (Issue 02/03), bukan konstanta global (lihat Task 1 & 7, K11).
- Metode isolir pelanggan static-IP (address-list default + alternatif) sebagai
  konfigurasi, dengan metode yang dipakai DISIMPAN agar resume deterministik
  (lihat Task 7).
- Sikap terhadap sesi aktif: setiap perubahan yang memutus/mengubah kelas layanan
  (PPPoE `disable`/`change_profile`, hotspot `disable`/`delete`, dan resume yang
  mengubah profil) WAJIB diikuti kill sesi aktif — kontrak driver, bukan usecase
  (K9). Issue ini hanya memastikan target yang benar direncanakan; sekuens
  [set/disable] → [/ppp active remove] (→ [cookie remove] untuk hotspot mode A)
  milik `internal/driver/mikrotik/commands.go`.
- Tiga endpoint REST: suspend / resume / terminate.

**Out of scope:**
- Pemrosesan baris sync ke perangkat nyata (milik Issue 01).
- Pengetahuan command MikroTik/GenieACS untuk disable/enable (driver — sudah ada
  di Issue 01/03; issue ini tidak menyentuh `internal/driver/`). Sekuens kill
  sesi (K9), penghapusan cookie hotspot (K9/K13), dan multi-path TR-069 (K13)
  adalah tanggung jawab driver, bukan issue ini.
- Auto-suspend terjadwal karena telat bayar (`system_scheduled`) — **Issue 14
  (Auto-Suspend & Auto-Restore Scheduler)**; issue ini hanya menyediakan usecase
  suspend/resume yang dipanggil ulang oleh scheduler itu dengan
  `actor_type='system_scheduled'`.
- Penambahan kolom `genieacs_device_id` (milik Issue 08).
- **Menyentuh OLT/ONU saat suspend/resume.** SUSPEND & RESUME **tidak pernah**
  mengirim command ke OLT — isolir dilakukan di lapisan MikroTik (PPPoE/hotspot)
  dan/atau GenieACS. Hanya **TERMINATE** yang boleh menyentuh OLT (`no onu <id>` /
  `ont delete`), dan itu operasi destruktif ber-HITL milik Issue 06 — issue ini
  hanya MENJADWALKAN target `olt_onu` `action=delete` bila subscription punya
  `onu_id`/`onu_pon_port` (Issue 05), bukan mengeksekusinya.

## REST API

Base path: `/api/v1`. Semua aksi ke perangkat mengikuti konvensi K4: handler TIDAK
memanggil driver langsung; usecase menulis baris `provisioning_sync_log`
`pending`, lalu response **202 Accepted** mengembalikan daftar id sync_log yang
dibuat (bukan hasil eksekusi perangkat).

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| POST | `/api/v1/subscriptions/:id/suspend` | Isolir layanan: cascade tulis baris sync disable/isolir ke semua target aktif | admin |
| POST | `/api/v1/subscriptions/:id/resume` | Aktifkan kembali layanan tersuspend: cascade enable/kembalikan profil | admin |
| POST | `/api/v1/subscriptions/:id/terminate` | Putus layanan permanen: cascade delete/hapus dari address-list | owner |

### POST `/api/v1/subscriptions/:id/suspend`

- **Request:** field penting `reason` (string, wajib — alasan suspend, disimpan
  ke `subscriptions.suspension_reason` dan `subscription_status_history.reason`).
- **Precondition:** status subscription saat ini harus `active`. Selain itu →
  gagal.
- **Response sukses (202):** objek berisi `subscription_id`, `status` baru
  (`suspended`), dan `sync_log_ids` (array UUID baris `provisioning_sync_log`
  yang dibuat, urut sesuai target). Sertakan juga `targets` (array ringkas:
  `target_type` + `action`) agar pemanggil tahu cascade apa yang dijadwalkan.
- **Response gagal:** `400` bila `reason` kosong; `404` bila subscription tidak
  ada; `409` bila status bukan `active` (mis. sudah `suspended`/`terminated`);
  `403` bila role kurang; `500` bila transaksi penulisan gagal.

### POST `/api/v1/subscriptions/:id/resume`

- **Request:** body kosong atau opsional `reason` (default `"manual resume"`).
- **Precondition:** status saat ini harus `suspended`.
- **Response sukses (202):** sama bentuk dengan suspend; `status` baru = `active`,
  `sync_log_ids` untuk aksi enable/kembalikan profil.
- **Response gagal:** `404` tidak ada; `409` bila status bukan `suspended`;
  `403` role kurang; `500` transaksi gagal.

### POST `/api/v1/subscriptions/:id/terminate`

- **Request:** field `reason` (string, wajib).
- **Precondition:** status saat ini `active` atau `suspended` (bukan
  `terminated`).
- **Response sukses (202):** `status` baru = `terminated`, `sync_log_ids` untuk
  aksi delete secret + hapus dari address-list (+ target `genieacs_tr069` bila
  fitur ACS diaktifkan & `genieacs_device_id` ada; + target `olt_onu`
  `action=delete` bila subscription punya `onu_id` — eksekusi OLT destruktif &
  ber-HITL, dijadwalkan di sini tapi dieksekusi Issue 06).
- **Response gagal:** `400` reason kosong; `404` tidak ada; `409` bila sudah
  `terminated`; `403` role kurang; `500` transaksi gagal.

## Tasks

**Task 1: Model transisi status & kebijakan isolir di domain subscription**

**Description:** Tambahkan aturan transisi status yang sah dan representasi
kebijakan isolir (disable vs change_profile) sebagai logika domain murni, tanpa
I/O.

**Acceptance criteria:**
- [ ] Ada fungsi validasi transisi (mis. `CanSuspend`, `CanResume`,
      `CanTerminate` pada `subscription.Subscription`) yang mengembalikan error
      sentinel bila transisi tidak sah (mis. `ErrNotActive`, `ErrNotSuspended`,
      `ErrAlreadyTerminated`).
- [ ] Ada tipe/enum kebijakan isolir MikroTik (mis. `IsolationPolicy` dengan
      nilai `ChangeProfile` sebagai **default** dan `DisableSecret` sebagai
      alternatif) di `internal/domain/subscription/`. Default `change_profile`
      sesuai K11 (profil ISOLIR membawa redirect portal + rate-limit + hook
      address-list; `disabled=yes` saja tidak).
- [ ] Nama **profil isolir** dan **profil normal (saat resume)** BUKAN konstanta
      global — dimodelkan sebagai nilai per-device/plan yang berasal dari
      plan/subscription (Issue 02/03). Domain menyediakan tipe/validasi bahwa
      kedua nama itu wajib ada saat mode `change_profile`; resume mengembalikan
      ke profil paket normal, bukan literal `default` (K11).
- [ ] Ada representasi **metode isolir static-IP** (mis.
      `MikrotikStaticSuspensionMethod` bernilai `address_list` sebagai default,
      plus `dhcp_block`/`bandwidth_limit`/`firewall_rule`) sebagai enum domain;
      metode yang dipakai per-subscription harus bisa disimpan agar resume
      deterministik (membalik metode yang sama yang dipakai saat suspend), lihat
      Task 7 (correction 5).
- [ ] Aturan **fail-safe**: bila mode `change_profile` tapi nama profil isolir
      tidak diketahui/kosong, fungsi domain mengembalikan error sentinel
      (mis. `ErrIsolationProfileMissing`) — suspend GAGAL eksplisit, tidak boleh
      diam-diam fallback ke `disable` atau profil lain (K11 fail-safe).
- [ ] Tidak ada import ke `adapter`/`driver`/framework eksternal (boundary
      domain, AGENTS.md §0).
- [ ] Sentinel error didefinisikan di `errors.go` domain subscription.

**Files likely touched:** `internal/domain/subscription/subscription.go`,
`internal/domain/subscription/errors.go`.

**Dependencies:** Issue 03 (domain subscription sudah ada).

**Estimated scope:** Small

---

**Task 2: Fungsi penentu target cascade (target planner)**

**Description:** Buat logika yang, dari satu `subscription.Subscription` + jenis
aksi (suspend/resume/terminate), menghasilkan DAFTAR target provisioning yang
harus ditulis — inilah inti temuan D.

**Acceptance criteria:**
- [ ] Ada fungsi (mis. `PlanSuspendTargets`, `PlanResumeTargets`,
      `PlanTerminateTargets`) yang mengembalikan slice deskriptor target
      (`target_type`, `action`, `device_id`) tanpa menyentuh DB.
- [ ] Aturan suspend terpenuhi: bila `pppoe_username` terisi → target
      `mikrotik_ppp_secret` dengan `action` sesuai kebijakan isolir Task 1
      (`change_profile` ke profil ISOLIR sebagai default, `disable` bila mode
      alternatif). **Kill sesi aktif WAJIB menyertai target ini** — baik
      `disable` MAUPUN `change_profile` (K9): perubahan `/ppp secret` tidak
      memutus sesi online, sekuens [set/disable] → [/ppp active remove] milik
      driver. Planner cukup memastikan target `mikrotik_ppp_secret` terjadwal;
      langkah kill adalah kontrak driver, bukan target sync terpisah.
      Bila `service_type='hotspot'` (voucher/hotspot terikat pelanggan, lihat
      Issue 13) → target `mikrotik_hotspot_user` `action=disable` (driver
      memasangkan kill semua sesi aktif + hapus cookie untuk mode `mac_login`,
      K9/K13); bila `static_ip` terisi → target `mikrotik_address_list`
      `action=create` (masuk address-list isolir); bila `genieacs_device_id`
      terisi **dan fitur ACS-suspend diaktifkan (feature-flag per plan/vendor)** →
      target `genieacs_tr069` `action=disable`. Default fitur ACS-suspend MATI.
- [ ] **Static vs PPPoE-dinamis untuk address-list isolir** dibedakan eksplisit
      (K11): untuk **static IP**, target `mikrotik_address_list` ditulis dengan
      IP konkret pelanggan. Untuk **PPPoE IP-dinamis**, keanggotaan address-list
      **tidak** ditulis manual saat suspend (IP sesi berikutnya bisa berbeda) —
      isolir dicapai lewat `change_profile` ke profil ISOLIR, dan keanggotaan
      address-list dipasang oleh **hook on-up di profil isolir** saat login
      berikutnya. Planner untuk PPPoE-dinamis TIDAK menghasilkan target
      `mikrotik_address_list` `create`.
- [ ] Aturan resume simetris: `mikrotik_ppp_secret` `action=enable` atau
      `change_profile` (kembali ke profil paket **normal dari plan/subscription**,
      bukan literal `default`), dengan kill sesi aktif juga menyertai bila
      profil berubah (K9); `mikrotik_hotspot_user` `action=enable`;
      `mikrotik_address_list` `action=delete` (keluarkan dari isolir, hanya untuk
      static IP — untuk PPPoE-dinamis, keanggotaan dilepas otomatis karena login
      berikutnya sudah di profil normal tanpa hook isolir); `genieacs_tr069`
      `action=enable` bila fitur ACS diaktifkan. Resume static-IP membalik
      **metode isolir yang tersimpan** dari suspend (Task 1/7), bukan menebak.
- [ ] Aturan terminate: `mikrotik_ppp_secret` `action=delete`,
      `mikrotik_hotspot_user` `action=delete` (+ kill semua sesi + hapus cookie
      mode `mac_login`), `mikrotik_address_list` `action=delete`, `genieacs_tr069`
      `action=disable` bila fitur ACS diaktifkan. **Hanya pada terminate**: bila
      subscription punya `onu_id`/`onu_pon_port` (Issue 05) → target `olt_onu`
      `action=delete` (destruktif, ber-HITL, dieksekusi Issue 06). SUSPEND &
      RESUME **tidak pernah** menghasilkan target `olt_onu`.
- [ ] Bila `genieacs_device_id` kosong ATAU fitur ACS-suspend nonaktif, target
      `genieacs_tr069` TIDAK dibuat (dilewati, bukan ditulis dengan device kosong).
- [ ] Kebijakan "minimal target terpenuhi" didokumentasikan sebagai komentar
      singkat + didefinisikan: minimal satu target MikroTik (secret, hotspot user,
      atau address-list) HARUS berhasil dijadwalkan agar transisi status dianggap
      valid; bila subscription tidak punya `pppoe_username`, bukan
      `service_type='hotspot'`, maupun `static_ip`, usecase mengembalikan error
      (tidak ada yang bisa diisolir). Target `genieacs_tr069`/`olt_onu` saja TIDAK
      memenuhi minimal (bukan jalur pemutus utama).
- [ ] Table-driven test mencakup: pppoe-dinamis (change_profile, tanpa
      address-list), pppoe-static (address-list create), hotspot-only,
      pppoe+genieacs (fitur ACS on/off), static+genieacs, hotspot+genieacs,
      terminate dengan onu_id (target olt_onu muncul), suspend dengan onu_id
      (target olt_onu TIDAK muncul), tanpa satu pun target MikroTik (error),
      kedua mode isolir.

**Files likely touched:** `internal/usecase/network/suspend_subscription.go`
(atau file planner terpisah di package sama), test di folder yang sama.

**Dependencies:** Task 1.

**Estimated scope:** Medium

---

**Task 3: Port yang dibutuhkan (repository subscription + status history + sync writer)**

**Description:** Pastikan kontrak port menyediakan operasi transaksional untuk
mengubah status subscription, menulis riwayat status, dan menulis banyak baris
sync sekaligus dalam satu transaksi.

**Acceptance criteria:**
- [ ] `internal/port/subscription_repository.go` punya method untuk memuat
      subscription by id dan mengubah status + timestamp terkait
      (`suspended_at`/`terminated_at`/`activated_at`) + `suspension_reason`.
- [ ] Ada kontrak penulisan `subscription_status_history` (bisa di port
      subscription atau port audit terpisah — nyatakan pilihan di PR) yang
      menerima `old_status`, `new_status`, `changed_by_user`,
      `changed_by_actor_type`, `reason`.
- [ ] Ada kontrak penulisan BATCH baris `provisioning_sync_log` (mis.
      `CreateSyncLogs(ctx, []SyncLogEntry) ([]uuid, error)`) yang menjamin
      penulisan atomik; interface ini di `internal/port/` (mis.
      `provisioning_sync_repository.go` bila belum dibuat Issue 01, atau reuse
      milik Issue 01).
- [ ] Ada satu operasi transaksional yang membungkus: update status +
      status_history + batch sync_log dalam SATU transaksi DB (atomik) —
      kontraknya diekspos lewat port (mis. metode `RunInTx` atau repository
      unit-of-work), bukan bocor ke usecase sebagai `*gorm.DB`.
- [ ] Semua method punya `context.Context` sebagai parameter pertama.

**Files likely touched:** `internal/port/subscription_repository.go`,
`internal/port/provisioning_sync_repository.go` (bila reuse dari Issue 01, cukup
dokumentasikan), kemungkinan `internal/port/audit_writer.go`.

**Dependencies:** Issue 01 (kontrak sync writer), Issue 03.

**Estimated scope:** Medium

---

**Task 4: Implementasi repository Postgres (transaksi atomik)**

**Description:** Implementasikan kontrak Task 3 di `internal/adapter/postgres/`,
menjamin update status + status_history + multi-baris sync_log commit bersama
atau rollback bersama.

**Acceptance criteria:**
- [ ] Implementasi transaksi tunggal: bila salah satu insert gagal, seluruh
      perubahan (status subscription, status_history, semua baris sync_log)
      di-rollback — tidak ada kondisi setengah jadi.
- [ ] Baris sync_log dibuat dengan `status='pending'`, `requested_at=now()`,
      `command_audit_log_id` NULL (diisi Sync Engine nanti).
- [ ] Baris status_history diisi `old_status`, `new_status`,
      `changed_by_actor_type='human'` untuk aksi via REST (untuk pemanggilan dari
      scheduler nanti, actor_type di-pass dari pemanggil).
- [ ] Error dibungkus `%w` dengan konteks operasi (AGENTS.md §4).
- [ ] Test integrasi pakai `testcontainers-go` (K7): verifikasi jumlah baris
      sync_log = jumlah target, status subscription berubah, satu baris
      status_history bertambah, dan rollback saat insert gagal.

**Files likely touched:** `internal/adapter/postgres/subscription_repository.go`,
mungkin `internal/adapter/postgres/provisioning_sync_repository.go`,
`internal/adapter/postgres/models.go`.

**Dependencies:** Task 3.

**Estimated scope:** Medium

---

**Task 5: Usecase suspend / resume / terminate (orkestrasi)**

**Description:** Rakit usecase orkestrasi yang memanggil target planner (Task 2),
memvalidasi transisi (Task 1), lalu menulis semua perubahan lewat repository
transaksional (Task 4), dan mengembalikan daftar id sync_log.

**Acceptance criteria:**
- [ ] `SuspendSubscription`, `ResumeSubscription`, `TerminateSubscription` di
      `internal/usecase/network/` (file terpisah per verb sesuai §1.4).
- [ ] Alur: muat subscription → validasi transisi (guard clause, early return) →
      hitung target cascade → tulis atomik (status + history + N baris sync
      pending) → kembalikan `[]uuid` sync_log + status baru.
- [ ] Usecase TIDAK memanggil `port.DeviceDriver` langsung (K4) — hanya menulis
      baris `pending`.
- [ ] Kebijakan minimal terpenuhi ditegakkan: bila planner mengembalikan nol
      target MikroTik yang valid, kembalikan error tanpa menulis apa pun.
- [ ] **Fail-safe profil isolir sebelum suspend (K11):** bila mode isolir
      `change_profile`, usecase memastikan nama profil isolir untuk device/plan
      subscription ini diketahui (dari plan/subscription, Issue 02/03) SEBELUM
      menulis baris sync. Bila tidak ada/kosong → suspend GAGAL eksplisit
      (kembalikan error, mis. `ErrIsolationProfileMissing`, map ke `409`/`422`),
      TIDAK menulis baris apa pun dan TIDAK diam-diam jatuh ke `disable`. (Verifikasi
      keberadaan profil di device secara nyata adalah tanggung jawab driver/Issue
      02 lewat "ensure profil isolir"; usecase menegakkan bahwa NAMA-nya wajib
      terdefinisi.)
- [ ] **Degradasi anggun saat device unreachable (K12):** urutan Polyglot = tulis
      baris sync (niat) lebih dulu; status subscription berubah saat baris
      **DITULIS**, bukan saat perangkat sukses (K4 async, kegagalan per-target
      ditangani per-baris oleh Sync Engine). Bila tidak ada target perangkat yang
      valid karena device unreachable / belum terpasang tetapi kebijakan bisnis
      menghendaki isolir tetap tercatat, usecase BOLEH mengubah status lokal saja
      **dengan** menuliskan `subscription_status_history.reason='local-only, device
      unreachable'` — tidak boleh diam-diam sukses tanpa jejak.
- [ ] Kebijakan isolir (mode + nama profil isolir/normal + metode static-IP)
      dibaca dari config/plan (Task 7), di-inject ke usecase, bukan hard-coded.
      Metode isolir static-IP yang dipakai DISIMPAN pada subscription/status agar
      resume membaliknya secara deterministik (Task 1/7).
- [ ] `context.Context` parameter pertama; error posisi terakhir; wrap `%w`.
- [ ] Table-driven test (mock repository) untuk tiap verb: happy path, transisi
      tidak sah, nol target, subscription tidak ditemukan, **profil isolir hilang
      saat mode change_profile (suspend gagal, nol baris ditulis)**, dan
      **jalur local-only saat tidak ada target device valid**.

**Files likely touched:** `internal/usecase/network/suspend_subscription.go`,
`internal/usecase/network/resume_subscription.go`,
`internal/usecase/network/terminate_subscription.go`, test di folder sama.

**Dependencies:** Task 1, 2, 4.

**Estimated scope:** Large

---

**Task 6: Handler REST + DTO + wiring route + RBAC**

**Description:** Ekspos ketiga usecase sebagai endpoint POST, dengan DTO
request/response, binding+validasi, dan pendaftaran route ber-RBAC.

**Acceptance criteria:**
- [ ] Tiga handler di `internal/adapter/http/subscription_handler.go` (reuse file
      handler subscription yang sudah ada dari Issue 03; tambah method, jangan
      buat file baru).
- [ ] DTO request/response di `internal/adapter/http/dto/` (mis.
      `suspend_subscription.go`): request `reason`, response `subscription_id` +
      `status` + `sync_log_ids` + `targets`.
- [ ] Validasi `reason` wajib untuk suspend & terminate → `400` bila kosong.
- [ ] Semua endpoint mengembalikan **202 Accepted** pada sukses (aksi ke
      perangkat asinkron, K4).
- [ ] Route didaftarkan di `internal/adapter/http/router.go` di bawah group
      `/api/v1` ber-middleware auth + RBAC.
- [ ] Mapping error usecase → status HTTP benar: `409` untuk transisi tidak sah,
      `404` not found, `403` role kurang, `400` validasi, `500` sisanya.
- [ ] Test handler pakai `httptest` (K7): sukses 202 tiap endpoint, 409 transisi
      salah, 400 reason kosong.

**Files likely touched:** `internal/adapter/http/subscription_handler.go`,
`internal/adapter/http/dto/suspend_subscription.go`,
`internal/adapter/http/router.go`.

**Dependencies:** Task 5.

**Estimated scope:** Medium

---

**Task 7: Konfigurasi kebijakan isolir + RBAC policy**

**Description:** Tambahkan opsi konfigurasi untuk kebijakan isolir dan pastikan
policy Casbin mengizinkan role yang benar untuk tiap aksi. Konfigurasi menetapkan
**default** dan sumber fallback; nama profil isolir/normal yang sebenarnya tetap
per-device/plan (Task 1, K11), bukan satu string global untuk seluruh sistem.

**Acceptance criteria:**
- [ ] `internal/config/config.go` punya opsi kebijakan isolir MikroTik:
      `MikrotikIsolationMode` bernilai `change_profile` (**default**, K11) atau
      `disable`; dan `MikrotikStaticSuspensionMethod` bernilai `address_list`
      (default) / `dhcp_block` / `bandwidth_limit` / `firewall_rule`. Default aman
      terdokumentasi.
- [ ] Ditegaskan bahwa **nama profil isolir & profil normal-resume BUKAN dari
      config global** melainkan dari plan/subscription (Issue 02/03); config hanya
      menyimpan mode + fallback opsional. Bila mode `change_profile` dan profil
      isolir tidak diketahui dari plan/subscription → suspend gagal (fail-safe
      Task 1/5), bukan diam-diam pakai nilai config global.
- [ ] Ada opsi **feature-flag GenieACS-suspend** (per plan/vendor, mis.
      `GenieacsSuspendWANEnabled`) dengan **default MATI** — cascade suspend
      default tidak menyentuh ACS (correction 7). Bila diaktifkan, target
      `genieacs_tr069 disable` baru direncanakan planner (Task 2).
- [ ] **Metode isolir static-IP yang dipakai per suspend DISIMPAN** (mis. kolom/
      field pada subscription atau dicatat di `status_history`/`sync_log` payload)
      agar resume membaliknya secara deterministik — bukan menebak ulang dari
      config yang mungkin sudah berubah sejak suspend.
- [ ] `configs/rbac_policy.csv` memuat entri: `suspend` & `resume` untuk
      superadmin/owner/admin (teknisi/staff tidak); `terminate` hanya untuk
      superadmin/owner.
- [ ] Nilai config di-inject ke usecase lewat konstruktor/wiring di
      `cmd/server/main.go`, bukan dibaca langsung dari environment di dalam
      usecase.

**Files likely touched:** `internal/config/config.go`, `configs/rbac_policy.csv`,
`cmd/server/main.go`.

**Dependencies:** Task 5 (usecase menerima config), Task 6 (route yang diproteksi).

**Estimated scope:** Small

---

**Task 8: Dokumentasi API + catatan cascade**

**Description:** Perbarui dokumentasi API dan tautkan bila membuat dokumen baru.

**Acceptance criteria:**
- [ ] `api/openapi.yaml` memuat tiga endpoint baru dengan request/response body
      dan kode status.
- [ ] Catatan singkat di dokumen yang relevan bahwa satu aksi suspend/terminate
      menghasilkan MULTI-BARIS `provisioning_sync_log` (rujuk temuan D).
- [ ] Bila menambah dokumen baru di root/docs, ditautkan dari `README.md` root
      pada commit yang sama (AGENTS.md §1.5).

**Files likely touched:** `api/openapi.yaml`.

**Dependencies:** Task 6.

**Estimated scope:** Small

---

**Task 9: Ensure isolir firewall/NAT infra per-router (prasyarat address-list)**

**Description:** Sediakan langkah idempoten yang memastikan infra isolir
(firewall drop `src-address-list=LIST_ISOLIR`, NAT redirect ke portal, allow
DNS/portal) ada di router SEBELUM address-list isolir dipakai. Tanpa infra ini,
memasukkan IP pelanggan ke address-list tidak memblokir apa pun (K11). Ini bukan
per-pelanggan melainkan one-time per-router.

**Acceptance criteria:**
- [ ] Ada usecase/target `mikrotik_firewall_isolir` (atau nama setara) yang
      menjadwalkan penulisan rule firewall+NAT isolir lewat `provisioning_sync_log`
      `pending` — tetap lewat K4, bukan panggil driver langsung.
- [ ] **Idempoten by comment** (K10/K11): driver menandai tiap rule dengan comment
      kanonik (mis. `poly:isolir-fw`) dan hanya membuat bila belum ada; menjalankan
      ulang tidak menduplikasi rule. Klasifikasi/urutan rule (drop di posisi yang
      benar, allow DNS/portal di atas drop) adalah pengetahuan driver
      (`internal/driver/mikrotik/commands.go`), bukan usecase.
- [ ] Dipasang idealnya saat setup/registrasi device (Issue 02); bila Issue 02
      belum menyediakannya, task ini bisa dipicu eksplisit (mis.
      `POST /api/v1/devices/:id/ensure-isolir-firewall`, role admin) ATAU dipanggil
      otomatis sekali sebelum suspend pertama yang memakai `mikrotik_address_list`.
      Nyatakan pilihan mekanisme di PR.
- [ ] Suspend yang bergantung address-list (static IP; atau hook profil isolir
      PPPoE dinamis) TIDAK dianggap "aman" bila infra ini belum ada — dokumentasikan
      keterkaitan ini agar tidak ada isolir yang bocor diam-diam.
- [ ] Test: idempotensi (dua kali jalan → satu set rule), dan verifikasi comment
      kanonik dipakai untuk deteksi keberadaan.

**Files likely touched:** `internal/usecase/network/` (usecase ensure-infra),
`internal/adapter/http/device_handler.go` (bila endpoint eksplisit),
`internal/driver/mikrotik/commands.go` (katalog rule — di luar issue ini bila
sudah ada; koordinasikan dengan Issue 01/02).

**Dependencies:** Issue 01 (sync engine), idealnya Issue 02 (setup device);
Task 2 (address-list sebagai target isolir).

**Estimated scope:** Medium

---

## Migrasi Database

Secara default **tidak ada perubahan skema**. Semua tabel yang dipakai sudah ada:
`subscriptions` (migrasi 000009), `subscription_status_history` (migrasi 000010),
`provisioning_sync_log` (migrasi 000011), `command_audit_log` (migrasi 000017).
Kolom `genieacs_device_id` di `subscriptions` ditambahkan oleh Issue 08 (bukan di
sini) — issue ini hanya membacanya bila sudah ada.

**Pengecualian bersyarat (metode isolir static-IP, correction 5):** kalau metode
isolir static-IP yang dipakai per-suspend (Task 7) diputuskan disimpan sebagai
**kolom baru** di `subscriptions` (bukan direkam di `status_history.reason` atau
payload `sync_log` yang sudah ada), itu butuh satu migrasi baru:
`000036_add_static_suspension_method_to_subscriptions` (sudah direservasi untuk
issue ini di tabel README §K6) — **tambahkan barisnya saat dipakai** pada PR yang
sama; jangan mengarang nomor di tengah urutan. Rekomendasi default: simpan di kolom
subscription agar resume deterministik dan bisa di-query, tetapi bila ingin
menghindari migrasi, rekam di `sync_log` payload/`status_history` (tanpa skema
baru). Bila memilih kolom, cerminkan juga di `DATABASE-SCHEMA.md` §6 (K6).

Selain kasus bersyarat di atas, tidak ada pembaruan `DATABASE-SCHEMA.md` yang
diperlukan selain (opsional) catatan naratif di §6.3 bahwa penulisan multi-baris
per event suspend/terminate adalah pola yang diharapkan.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/domain/subscription/...` — validasi transisi & kebijakan
      isolir hijau.
- [ ] `go test ./internal/usecase/network/...` — table-driven suspend/resume/
      terminate + target planner hijau.
- [ ] `go test ./internal/adapter/postgres/...` — integrasi transaksi atomik
      (testcontainers) hijau: jumlah baris sync = jumlah target, rollback saat
      gagal.
- [ ] `go test ./internal/adapter/http/...` — handler httptest 202/409/400 hijau.
- [ ] `make lint` bersih (gofumpt, staticcheck, boundary import).
- [ ] Smoke test manual: `curl -X POST` ke `/api/v1/subscriptions/:id/suspend`
      dengan body `reason` dan header `Authorization: Bearer <token admin>` →
      cek response 202 berisi `sync_log_ids`, lalu `curl` GET subscription →
      status `suspended`, dan query `provisioning_sync_log` menunjukkan
      >1 baris `pending` untuk subscription itu (pppoe + genieacs bila ada).
- [ ] Smoke test resume & terminate simetris; verifikasi 409 saat suspend
      subscription yang sudah `suspended`.
- [ ] Verifikasi boundary: suspend/resume subscription ber-`onu_id` TIDAK
      menghasilkan baris `olt_onu`; hanya terminate yang menghasilkannya.
- [ ] Verifikasi fail-safe: suspend saat mode `change_profile` tetapi profil
      isolir tidak diketahui dari plan/subscription → gagal eksplisit, nol baris
      `sync_log` ditulis, status subscription tetap `active`.
- [ ] Verifikasi default GenieACS: tanpa mengaktifkan feature-flag ACS-suspend,
      suspend TIDAK menulis baris `genieacs_tr069` meski `genieacs_device_id` ada.
- [ ] (Bila Task 9 diaktifkan) verifikasi infra isolir firewall dipasang idempoten
      — dijalankan dua kali menghasilkan satu set rule ber-comment kanonik.

## Definition of Done

- [ ] Tiga endpoint suspend/resume/terminate jalan, mengembalikan 202 +
      `sync_log_ids`.
- [ ] Satu event menghasilkan MULTI-BARIS `provisioning_sync_log` `pending`
      ditulis atomik bersama update status + `subscription_status_history`
      (temuan D / #9 tertutup).
- [ ] Kegagalan sebagian tercermin per-baris oleh Sync Engine (Issue 01); status
      subscription hanya berubah bila kebijakan minimal target terpenuhi.
- [ ] Cascade GenieACS bersifat opsional (feature-flag per plan/vendor, default
      MATI); bila diaktifkan, ditulis hanya saat `genieacs_device_id` ada,
      menggunakan path native `WAN*Connection.1.Enable` yang diuji per vendor (K13).
- [ ] SUSPEND/RESUME tidak pernah menyentuh OLT; hanya TERMINATE menjadwalkan
      target `olt_onu delete` (destruktif, ber-HITL, eksekusi Issue 06).
- [ ] Kill sesi aktif dipastikan menyertai target pemutus/pengubah profil (PPPoE
      `disable`/`change_profile` DAN resume yang mengubah profil, hotspot
      `disable`/`delete` + hapus cookie mode `mac_login`) — kontrak driver (K9),
      dites di Issue 01/driver; issue ini memastikan target yang benar direncanakan.
- [ ] Isolir utama = `change_profile` ke profil ISOLIR per-device/plan dengan
      fail-safe (profil isolir wajib diketahui, kalau tidak suspend gagal);
      static-IP pakai address-list dengan metode tersimpan agar resume deterministik
      (K11).
- [ ] Degradasi anggun terdukung: bila device unreachable, status boleh berubah
      lokal dengan `status_history.reason='local-only, device unreachable'` (K12).
- [ ] RBAC benar: admin+ untuk suspend/resume, owner untuk terminate; staff/
      teknisi ditolak.
- [ ] Kebijakan isolir (disable vs change_profile) dikonfigurasi, bukan
      hard-coded.
- [ ] Handler tidak memanggil `port.DeviceDriver` langsung (K4 dipatuhi).
- [ ] Semua test hijau, `make lint` bersih, `api/openapi.yaml` diperbarui.
