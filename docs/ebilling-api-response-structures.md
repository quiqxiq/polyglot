# Struktur Respons API e-Billing

Dokumen ini menjelaskan **bentuk/struktur JSON** respons API e-Billing, bukan nilai
datanya. Fokus utama: `GET /pelanggan` dan `GET /tipe-pembayaran`, dilengkapi
endpoint GET lain yang sudah diuji.

- **Base URL:** `https://api-service.ebilling.id/api`
- **Sumber struktur:** respons aktual tanggal 30 Oktober 2026 + OpenAPI spec
- **Sifat:** read-only; tidak ada request tulis yang dilakukan
- **Nilai data:** tidak ditampilkan; hanya nama field, tipe, nullable/opsional,
  dan nesting

Notasi tipe:

| Notasi | Arti |
|---|---|
| `string` | JSON string |
| `integer` | JSON number tanpa desimal |
| `number` | JSON number, bisa desimal |
| `boolean` | `true` / `false` |
| `string (date)` | string berformat tanggal `YYYY-MM-DD` |
| `string (datetime)` | string berformat tanggal-waktu |
| `T[]` | array berisi `T` |
| `T \| null` | bisa `T` atau `null` |
| `null` | selalu `null` pada sampel yang diamati |
| `?` | opsional; tidak selalu muncul |

---

## 1. Pola umum respons

### 1.1 Collection (daftar berpaginasi)

Sebagian besar endpoint daftar memakai bentuk:

```ts
type Collection<T> = {
  data: T[]
  links: Links
  meta: Meta
}
```

```ts
type Links = {
  first: string | null
  last: string | null
  prev: string | null
  next: string | null
}
```

```ts
type Meta = {
  current_page: number
  from: number | null
  last_page: number
  links: MetaLink[]
  path: string
  per_page: number
  to: number | null
  total: number
}

type MetaLink = {
  url: string | null
  label: string
  page: number | null
  active: boolean
}
```

Catatan:

- `links.first`, `links.last`, `links.prev`, `links.next` berisi URL atau `null`.
- `meta.from` dan `meta.to` bisa `null` jika halaman kosong.
- `meta.links[]` adalah link paginasi per halaman.
- Pola ini berasal dari paginator Laravel.

### 1.2 Single resource

Endpoint yang mengembalikan satu objek biasanya memakai:

```ts
type Item<T> = {
  data: T
}
```

Contoh: `GET /me`.

### 1.3 Flat object

Beberapa endpoint mengembalikan objek langsung tanpa `data`:

```ts
type Flat = {
  // field-field langsung
}
```

Contoh: `GET /dashboard-mitra`, `GET /pelanggan/pppoe/{id_pelanggan}`,
`GET /hello-world`, `GET /health`.

### 1.4 Error

Bentuk error validasi:

```ts
type ValidationError = {
  message: string
  errors?: Record<string, string[]>
}
```

Bentuk error umum (401/403/404/429) biasanya hanya:

```ts
type ErrorResponse = {
  message: string
}
```

---

## 2. Tipe `Customer`

Dipakai oleh:

- `GET /pelanggan` → `Collection<Customer>`
- `GET /me` → `Item<Customer>`
- `GET /resellers/{account_reseller}/customers` → `Collection<Customer>`
- `POST /pelanggan` → `Item<Customer>` (menurut spec; tidak dipanggil)

```ts
type Customer = {
  id_warga: number
  id_pelanggan: string
  tgl_registrasi: string
  account: string
  id_lokasi: string
  nama_warga: string
  id_tipe_pembayaran: string
  nama_tipe: string
  keterangan: string
  harga: number
  ppn: number | null
  nama_lain_lain: string | null
  nominal_lain_lain: number | null
  jns_tipe_pembayaran: string
  username: string
  alamat: string
  titik_lokasi: string | null
  blok: string
  no_rumah: string
  rt: string
  rw: string
  dusun: string
  kelurahan: string
  kecamatan: string
  kabupaten: string
  propinsi: string
  kode_pos: string
  tlp: string
  tlp2: string | null
  email: string
  level: string
  status: string
  status_langganan: string
  catatan: string | null
  id_sales: number | null
  nama_sales: string | null
  id_router: string | null
  jns_koneksi: string | null
  tgl_jatuh_tempo: string | null
  prorate: string | null
  jns_blokir: string | null
  latitude: string | null
  longitude: string | null
  modul_olt: string | null
  olt_port: string | null
  index_onu: string | null
  sn_onu: string | null
  insentif_sales: string | null
  metode_insentif: string | null
  nominal_insentif: string | null
  id_olt: null
  id_odc: null
  id_odp: null
  nama_olt: string | null
  nama_odc: null
  nama_odp: null
  status_koneksi: string | null
  waktu_pengecekan: string | null
  jns_tagihan: string | null
  tgl_tagihan: string | null
  tgl_isolir: string | null
}
```

### 2.1 Keterangan field `Customer`

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `id_warga` | `integer` | tidak | ID internal pelanggan |
| `id_pelanggan` | `string` | tidak | Kode pelanggan |
| `tgl_registrasi` | `string (date)` | tidak | Tanggal registrasi |
| `account` | `string` | tidak | Kode tenant |
| `id_lokasi` | `string` | tidak | ID lokasi; **string**, bukan integer |
| `nama_warga` | `string` | tidak | Nama pelanggan |
| `id_tipe_pembayaran` | `string` | tidak | ID paket; **string**, bukan integer |
| `nama_tipe` | `string` | tidak | Nama paket |
| `keterangan` | `string` | tidak | Keterangan paket |
| `harga` | `integer` | tidak | Harga langganan |
| `ppn` | `integer` | ya | Nilai PPN; bisa `null` pada profil admin |
| `nama_lain_lain` | `string` | ya | Nama biaya tambahan |
| `nominal_lain_lain` | `integer` | ya | Nominal biaya tambahan |
| `jns_tipe_pembayaran` | `string` | tidak | Jenis paket |
| `username` | `string` | tidak | Username pelanggan |
| `alamat` | `string` | tidak | Alamat |
| `titik_lokasi` | `string` | ya | Titik lokasi; `null` pada `/me` |
| `blok` | `string` | tidak | Blok |
| `no_rumah` | `string` | tidak | Nomor rumah |
| `rt` | `string` | tidak | RT |
| `rw` | `string` | tidak | RW |
| `dusun` | `string` | tidak | Dusun |
| `kelurahan` | `string` | tidak | Kelurahan |
| `kecamatan` | `string` | tidak | Kecamatan |
| `kabupaten` | `string` | tidak | Kabupaten |
| `propinsi` | `string` | tidak | Provinsi |
| `kode_pos` | `string` | tidak | Kode pos |
| `tlp` | `string` | tidak | Telepon utama |
| `tlp2` | `string` | ya | Telepon kedua |
| `email` | `string` | tidak | Email; bisa string kosong |
| `level` | `string` | tidak | Level pengguna |
| `status` | `string` | tidak | Kode status akun; **string**, bukan integer |
| `status_langganan` | `string` | tidak | Status langganan |
| `catatan` | `string` | ya | Catatan |
| `id_sales` | `integer` | ya | ID sales |
| `nama_sales` | `string` | ya | Nama sales |
| `id_router` | `string` | ya | ID router |
| `jns_koneksi` | `string` | ya | Jenis koneksi |
| `tgl_jatuh_tempo` | `string` | ya | Hari jatuh tempo; **string** |
| `prorate` | `string` | ya | Nilai prorate |
| `jns_blokir` | `string` | ya | Jenis blokir |
| `latitude` | `string` | ya | Latitude |
| `longitude` | `string` | ya | Longitude |
| `modul_olt` | `string` | ya | Modul OLT |
| `olt_port` | `string` | ya | Port OLT |
| `index_onu` | `string` | ya | Index ONU |
| `sn_onu` | `string` | ya | Serial number ONU |
| `insentif_sales` | `string` | ya | Insentif sales |
| `metode_insentif` | `string` | ya | Metode insentif |
| `nominal_insentif` | `string` | ya | Nominal insentif |
| `id_olt` | `null` | ya | Tidak terisi pada sampel |
| `id_odc` | `null` | ya | Tidak terisi pada sampel |
| `id_odp` | `null` | ya | Tidak terisi pada sampel |
| `nama_olt` | `string` | ya | Nama OLT |
| `nama_odc` | `null` | ya | Tidak terisi pada sampel |
| `nama_odp` | `null` | ya | Tidak terisi pada sampel |
| `status_koneksi` | `string` | ya | Status koneksi |
| `waktu_pengecekan` | `string (datetime)` | ya | Waktu pengecekan |
| `jns_tagihan` | `string` | ya | Jenis tagihan |
| `tgl_tagihan` | `string (date)` | ya | Tanggal tagihan |
| `tgl_isolir` | `string (date)` | ya | Tanggal isolir |

Catatan: `ppn`, `nama_lain_lain`, `nominal_lain_lain`, dan `titik_lokasi`
terisi pada sampel `GET /pelanggan`, tetapi bernilai `null` pada `GET /me`.
Tipe di atas menggabungkan kedua endpoint; untuk parsing gunakan tipe nullable.

### 2.2 Bentuk respons `GET /pelanggan`

```ts
type GetPelangganResponse = {
  data: Customer[]
  links: Links
  meta: Meta
}
```

Query yang tersedia: `page`, `per_page`, `id_pelanggan`, `id_lokasi`.

### 2.3 Bentuk respons `GET /me`

```ts
type GetMeResponse = {
  data: Customer
}
```

---

## 3. Tipe `Package` (`GET /tipe-pembayaran`)

Endpoint paket mengembalikan `Collection<Package>`.

```ts
type Package = {
  id_tipe_pembayaran: number
  account: number
  nama_tipe: string
  keterangan: string
  harga: number
  jns_tipe_pembayaran: string
  id_lokasi: number
  profile: string | null
  registrasi: string
  filter_lokasi: string
  fee: null
  wkt_update: null
  id_area: null
  sub_area: null
  nama_area: null
  nama_cabang: null
  id_cabang: null
}
```

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `id_tipe_pembayaran` | `integer` | tidak | ID paket |
| `account` | `integer` | tidak | Kode tenant |
| `nama_tipe` | `string` | tidak | Nama paket |
| `keterangan` | `string` | tidak | Keterangan paket |
| `harga` | `integer` | tidak | Harga paket |
| `jns_tipe_pembayaran` | `string` | tidak | Jenis paket |
| `id_lokasi` | `integer` | tidak | ID lokasi |
| `profile` | `string` | ya | Profil paket |
| `registrasi` | `string` | tidak | Penanda registrasi |
| `filter_lokasi` | `string` | tidak | Filter lokasi |
| `fee` | `null` | ya | Tidak terisi pada sampel |
| `wkt_update` | `null` | ya | Tidak terisi pada sampel |
| `id_area` | `null` | ya | Tidak terisi pada sampel |
| `sub_area` | `null` | ya | Tidak terisi pada sampel |
| `nama_area` | `null` | ya | Tidak terisi pada sampel |
| `nama_cabang` | `null` | ya | Tidak terisi pada sampel |
| `id_cabang` | `null` | ya | Tidak terisi pada sampel |

Bentuk respons:

```ts
type GetTipePembayaranResponse = {
  data: Package[]
  links: Links
  meta: Meta
}
```

Query: `page`, `per_page`.

---

## 4. Tipe `Location` (`GET /lokasi`)

```ts
type Location = {
  id_lokasi: number
  account: number
  nama_lokasi: string
  alamat_lokasi: string
  tlp_lokasi: string
  group_wa: string
  id_pic: null
  nama_pic: null
  kode_lokasi: null
  account_wagw: null
  insentif_sales: null
  metode_insentif: null
  nominal_insentif: null
  filter_lokasi: string
  jns_lokasi: string
  id_cabang: null
  nama_cabang: null
  id_referensi_corcab: null
  nama_referensi_corcab: null
  provinsi: null
  kabupaten: null
  kecamatan: null
  kelurahan: null
  rt: null
  rw: null
}
```

Bentuk respons: `Collection<Location>`. Query: `page`, `per_page`.

---

## 5. Tipe `BankAccount` (`GET /rekening`)

```ts
type BankAccount = {
  id_rekening: number
  account: string
  nama_rekening: string
  atas_nama: string
  no_rekening: string
  invoice: string
  invoice_corporate: string
  id_lokasi: null
  nama_lokasi: null
}
```

Bentuk respons: `Collection<BankAccount>`. Query: `page`, `per_page`,
`account`, `nama_rekening`, `atas_nama`, `no_rekening`.

---

## 6. Tipe `Sales` (`GET /sales`)

```ts
type Sales = {
  id_warga: number
  nama_warga: string
  level: string
  status: string
}
```

Bentuk respons: `Collection<Sales>`. Query: `page`, `per_page`.

---

## 7. Tipe `Payment` (`GET /pembayaran-pelanggan`)

Dipakai juga oleh `GET /resellers/{account_reseller}/payments`.

```ts
type Payment = {
  id_ipl: number
  id_pelanggan: string | null
  nama_pelanggan: string
  nama_sales: string
  nama_pembayaran: string
  nominal_harus_dibayar: number
  nominal_pembayaran: number
  status_pembayaran: null
  alamat: string
  tlp: string
  lokasi: number
  bukti_pembayaran: string
  periode_pembayaran: string
  metode_pembayaran: string
  nama_rekening: string
  atas_nama: string
  no_rekening: string
  waktu_entry: string
  keterangan: string
  metode_insentif: string
  insentif: string
  nominal_insentif: string
}
```

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `id_ipl` | `integer` | tidak | ID transaksi pembayaran |
| `id_pelanggan` | `string` | ya | Kode pelanggan |
| `nama_pelanggan` | `string` | tidak | Nama pelanggan |
| `nama_sales` | `string` | tidak | Nama sales |
| `nama_pembayaran` | `string` | tidak | Nama paket/pembayaran |
| `nominal_harus_dibayar` | `integer` | tidak | Nominal tagihan |
| `nominal_pembayaran` | `integer` | tidak | Nominal dibayar |
| `status_pembayaran` | `null` | ya | Tidak terisi pada sampel |
| `alamat` | `string` | tidak | Alamat |
| `tlp` | `string` | tidak | Telepon |
| `lokasi` | `integer` | tidak | ID lokasi; **integer** di endpoint ini |
| `bukti_pembayaran` | `string` | tidak | Nama/file bukti pembayaran |
| `periode_pembayaran` | `string (date)` | tidak | Periode pembayaran |
| `metode_pembayaran` | `string` | tidak | Metode pembayaran |
| `nama_rekening` | `string` | tidak | Nama rekening/metode |
| `atas_nama` | `string` | tidak | Nama pemilik rekening |
| `no_rekening` | `string` | tidak | Nomor rekening |
| `waktu_entry` | `string (datetime)` | tidak | Waktu pencatatan |
| `keterangan` | `string` | tidak | Keterangan |
| `metode_insentif` | `string` | tidak | Metode insentif |
| `insentif` | `string` | tidak | Nilai insentif |
| `nominal_insentif` | `string` | tidak | Nominal insentif |

Bentuk respons:

```ts
type GetPembayaranResponse = {
  data: Payment[]
  links: Links
  meta: Meta
}
```

Query: `page`, `per_page`, `from`, `to`, `bulan`, `id_pelanggan`, `id_lokasi`.

---

## 8. Tipe `UnpaidCustomer` (`GET /pelanggan-belum-bayar`)

```ts
type UnpaidCustomer = {
  id_warga: number
  id_pelanggan: string
  nama_pelanggan: string
  tlp: string
  alamat: string
  nama_langganan: string
  keterangan: string
  nominal_tagihan: number
  lokasi: string
  status_langganan: string
  tgl_registrasi: string
  tgl_jatuh_tempo: string
  periode_terakhir_dibayar: string
  periode_belum_lunas: UnpaidPeriod[]
  jumlah_periode_belum_lunas: number
  total_nominal_belum_lunas: number
}
```

```ts
type UnpaidPeriod = {
  periode: string
  label_periode: string
  nominal: number
  status_pembayaran: string
}
```

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `id_warga` | `integer` | tidak | ID internal pelanggan |
| `id_pelanggan` | `string` | tidak | Kode pelanggan |
| `nama_pelanggan` | `string` | tidak | Nama pelanggan |
| `tlp` | `string` | tidak | Telepon |
| `alamat` | `string` | tidak | Alamat |
| `nama_langganan` | `string` | tidak | Nama paket |
| `keterangan` | `string` | tidak | Keterangan paket |
| `nominal_tagihan` | `integer` | tidak | Nominal tagihan |
| `lokasi` | `string` | tidak | ID lokasi; **string** di endpoint ini |
| `status_langganan` | `string` | tidak | Status langganan |
| `tgl_registrasi` | `string (date)` | tidak | Tanggal registrasi |
| `tgl_jatuh_tempo` | `string` | tidak | Hari jatuh tempo |
| `periode_terakhir_dibayar` | `string` | tidak | Label periode terakhir dibayar |
| `periode_belum_lunas` | `UnpaidPeriod[]` | tidak | Daftar periode belum lunas |
| `jumlah_periode_belum_lunas` | `integer` | tidak | Jumlah periode belum lunas |
| `total_nominal_belum_lunas` | `integer` | tidak | Total nominal belum lunas |

Field `periode_belum_lunas[]`:

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `periode` | `string (date)` | tidak | Tanggal/periode tagihan |
| `label_periode` | `string` | tidak | Label periode untuk tampilan |
| `nominal` | `integer` | tidak | Nominal periode |
| `status_pembayaran` | `string` | tidak | Status pembayaran periode |

Bentuk respons:

```ts
type GetPelangganBelumBayarResponse = {
  data: UnpaidCustomer[]
  links: Links
  meta: Meta
}
```

Query: `page`, `per_page`, `sampai`, `id_lokasi`, `tgl_jatuh_tempo`.

---

## 9. Tipe `StatusRow` (`GET /status-pelanggan`)

```ts
type StatusRow = {
  id_warga: number
  id_pelanggan: string
  nama_warga: string
  account: string
  status: string
  status_langganan: string
  status_pelanggan: string
}
```

Bentuk respons: `Collection<StatusRow>`. Query: `page`, `per_page`, `id_warga`,
`status_langganan`.

---

## 10. Tipe `ExpiredRow` (`GET /expired`)

```ts
type ExpiredRow = {
  id_pelanggan: string
  nama_tipe: string
  pppoe_user: string
  pppoe_password: string
  nama_pelanggan: string
  status: string
  expired_at: string
  invoice: string
}
```

| Field | Tipe | Nullable | Keterangan |
|---|---|---|---|
| `id_pelanggan` | `string` | tidak | Kode pelanggan |
| `nama_tipe` | `string` | tidak | Nama paket |
| `pppoe_user` | `string` | tidak | Username PPPoE |
| `pppoe_password` | `string` | tidak | Password PPPoE |
| `nama_pelanggan` | `string` | tidak | Nama pelanggan |
| `status` | `string` | tidak | Status |
| `expired_at` | `string` | tidak | Tanggal kedaluwarsa; sering kosong |
| `invoice` | `string` | tidak | URL invoice |

Bentuk respons `GET /expired` **berbeda** dari collection biasa:

```ts
type GetExpiredResponse = {
  data: ExpiredRow[]
  links: Links
  meta: ExpiredMeta
}

type ExpiredMeta = {
  current_page: number
  current_page_url: string
  from: number | null
  path: string
  per_page: number
  to: number | null
}
```

Catatan: `meta` pada `/expired` tidak memiliki `last_page`, `total`, atau
`links`; `links.last` juga bisa `null`. Paginasi ini bersifat sederhana.

Query: `page`, `per_page`, `search`.

---

## 11. Tipe `Reseller` (`GET /resellers`)

```ts
type Reseller = {
  id_mitra: number
  account_reseller: string
  expired: string
  nama_reseller: string
  nama_pic: string
  alamat: string
  tlp: string
  email: string | null
  bulan: string

  total_pelanggan: number
  total_lunas: number
  total_belum_lunas: number
  total_pelanggan_ppn: number
  nominal_pelanggan_ppn: number
  total_pelanggan_non_ppn: number
  nominal_pelanggan_non_ppn: number

  total_pemasukan_ppn: number
  total_ppn_didapat: number
  total_pemasukan_non_ppn: number
  total_pemasukan: number
  total_revenue: number

  pengaturan_bhp: null
  total_bhp: number
  pengaturan_uso: null
  total_uso: number
  total_pemasukan_setelah_ppn_bhp_uso: number

  metode_fee: null
  nominal_fee: null
  total_fee_isp: number
  total_estimasi_pemasukan: number
}
```

Kelompok field:

| Kelompok | Field |
|---|---|
| Identitas | `id_mitra`, `account_reseller`, `expired`, `nama_reseller`, `nama_pic`, `alamat`, `tlp`, `email`, `bulan` |
| Pelanggan | `total_pelanggan`, `total_lunas`, `total_belum_lunas`, `total_pelanggan_ppn`, `nominal_pelanggan_ppn`, `total_pelanggan_non_ppn`, `nominal_pelanggan_non_ppn` |
| Pemasukan | `total_pemasukan_ppn`, `total_ppn_didapat`, `total_pemasukan_non_ppn`, `total_pemasukan`, `total_revenue` |
| BHP/USO | `pengaturan_bhp`, `total_bhp`, `pengaturan_uso`, `total_uso`, `total_pemasukan_setelah_ppn_bhp_uso` |
| Fee/estimasi | `metode_fee`, `nominal_fee`, `total_fee_isp`, `total_estimasi_pemasukan` |

Bentuk respons: `Collection<Reseller>`. Query: `page`, `per_page`, `bulan`.

---

## 12. Tipe `Dashboard` (`GET /dashboard-mitra`)

```ts
type Dashboard = {
  bulan: string
  total_revenue: number
  total_pelanggan: number
  total_reseller: number
  active_customer: number
  non_active_customer: number
  persen_active: number
  persen_non_active: number
}
```

Bentuk respons: flat object. Query: `bulan`.

---

## 13. Tipe `PPPoEInfo` (`GET /pelanggan/pppoe/{id_pelanggan}`)

```ts
type PPPoEInfo = {
  id_pelanggan: string
  nama_warga: string
  nama_lokasi: string
  nama_router: string
  user_pppoe: string
  password_pppoe: string
}
```

Bentuk respons: flat object, **tanpa** `data`/`links`/`meta`.

Catatan: path `{id_pelanggan}` memakai ID numerik yang sama dengan `id_warga`
pada `GET /pelanggan`, bukan kode `id_pelanggan`.

---

## 14. Tipe `InstalasiPelangganBaru` (`GET /instalasi-pelanggan-baru`)

Bentuk respons:

```ts
type GetInstalasiResponse = {
  data: unknown[]
  links: Links
  meta: Meta
}
```

Pada sampel, `data` kosong, sehingga struktur item belum bisa diamati. Query:
`page`, `per_page`.

---

## 15. Tipe `HelloWorld` (`GET /hello-world`)

```ts
type HelloWorld = {
  message: string
  account: string
  id_warga_legacy: number
}
```

Bentuk respons: flat object.

---

## 16. Tipe `Health` (`GET /health`)

```ts
type Health = {
  ok: boolean
  message: string
}
```

Bentuk respons: flat object.

---

## 17. Tipe `LoginResponse` (`POST /login`)

Tidak dipanggil karena bersifat tulis. Menurut spec:

```ts
type LoginResponse = {
  token: string
  warga: object
}
```

`warga` kemungkinan besar memakai bentuk `Customer`, tetapi belum diverifikasi
karena endpoint login tidak dipanggil.

---

## 18. Peta endpoint → tipe respons

| Endpoint | Tipe respons |
|---|---|
| `GET /health` | `Health` |
| `GET /hello-world` | `HelloWorld` |
| `GET /me` | `Item<Customer>` |
| `GET /dashboard-mitra` | `Dashboard` |
| `GET /resellers` | `Collection<Reseller>` |
| `GET /resellers/{account_reseller}/customers` | `Collection<Customer>` |
| `GET /resellers/{account_reseller}/payments` | `Collection<Payment>` |
| `GET /lokasi` | `Collection<Location>` |
| `GET /tipe-pembayaran` | `Collection<Package>` |
| `GET /rekening` | `Collection<BankAccount>` |
| `GET /sales` | `Collection<Sales>` |
| `GET /instalasi-pelanggan-baru` | `Collection<unknown>` |
| `GET /pelanggan` | `Collection<Customer>` |
| `POST /pelanggan` | `Item<Customer>` (spec; tidak dipanggil) |
| `GET /pelanggan-belum-bayar` | `Collection<UnpaidCustomer>` |
| `GET /status-pelanggan` | `Collection<StatusRow>` |
| `GET /expired` | `Collection<ExpiredRow>` dengan `ExpiredMeta` |
| `GET /pembayaran-pelanggan` | `Collection<Payment>` |
| `POST /pembayaran-pelanggan` | `Item<Payment>` (spec; tidak dipanggil) |
| `POST /pembayaran-pelanggan-2809` | `Item<Payment>` (spec; tidak dipanggil) |
| `GET /pelanggan/pppoe/{id_pelanggan}` | `PPPoEInfo` |
| `POST /login` | `LoginResponse` (spec; tidak dipanggil) |
| `POST /logout` | `204 No Content` |

---

## 19. Catatan penting untuk parsing

1. **Tipe JSON tidak selalu sama dengan tipe semantik.**
   - `id_lokasi` pada `Customer` adalah `string`, tetapi pada `Package` dan
     `Location` adalah `integer`.
   - `lokasi` pada `Payment` adalah `integer`, tetapi pada `UnpaidCustomer`
     adalah `string`.
   - `status` pada `Customer`/`StatusRow` adalah `string`, bukan integer.
   - `tgl_jatuh_tempo` adalah `string`, bukan integer.
2. **Banyak field numerik dikirim sebagai string.** Contohnya
   `id_tipe_pembayaran`, `id_lokasi`, dan `nominal_insentif` pada `Customer`.
3. **Field yang selalu `null` pada sampel** belum tentu tidak akan pernah
   terisi. Perlakukan sebagai nullable/unknown:
   - `Customer.id_olt`, `id_odc`, `id_odp`, `nama_odc`, `nama_odp`
   - `Package.fee`, `wkt_update`, `id_area`, `sub_area`, `nama_area`,
     `nama_cabang`, `id_cabang`
   - `Payment.status_pembayaran`
   - `Reseller.pengaturan_bhp`, `pengaturan_uso`, `metode_fee`, `nominal_fee`
4. **`/expired` memakai meta paginasi berbeda.** Jangan memakai parser `Meta`
   yang sama dengan endpoint daftar lain.
5. **`/dashboard-mitra`, `/pelanggan/pppoe/{id}`, `/hello-world`, dan `/health`
   tidak memakai envelope `data`.**
6. **`/me` memakai `{ data: Customer }`, bukan array.**
7. **`/resellers/{account}/payments` memakai bentuk `Payment` yang sama**
   dengan `/pembayaran-pelanggan`, tetapi nilai `meta` dan `links` mengikuti
   collection biasa.
8. **`/pelanggan-belum-bayar` memiliki nested array** `periode_belum_lunas`
   yang perlu diparse sebagai `UnpaidPeriod[]`, bukan string.
9. **`/expired` memuat password PPPoE dan URL invoice.** Jangan log seluruh
   objek; pilih field yang diperlukan saja.
10. **`/rekening` memuat nomor rekening dan nama pemilik.** Masking diperlukan
    saat data ditampilkan atau disimpan.

---

## 20. Ringkasan bentuk JSON tanpa nilai

### `GET /pelanggan`

```json
{
  "data": [
    {
      "id_warga": "<integer>",
      "id_pelanggan": "<string>",
      "tgl_registrasi": "<date>",
      "account": "<string>",
      "id_lokasi": "<string>",
      "nama_warga": "<string>",
      "id_tipe_pembayaran": "<string>",
      "nama_tipe": "<string>",
      "keterangan": "<string>",
      "harga": "<integer>",
      "ppn": "<integer|null>",
      "nama_lain_lain": "<string|null>",
      "nominal_lain_lain": "<integer|null>",
      "jns_tipe_pembayaran": "<string>",
      "username": "<string>",
      "alamat": "<string>",
      "titik_lokasi": "<string|null>",
      "blok": "<string>",
      "no_rumah": "<string>",
      "rt": "<string>",
      "rw": "<string>",
      "dusun": "<string>",
      "kelurahan": "<string>",
      "kecamatan": "<string>",
      "kabupaten": "<string>",
      "propinsi": "<string>",
      "kode_pos": "<string>",
      "tlp": "<string>",
      "tlp2": "<string|null>",
      "email": "<string>",
      "level": "<string>",
      "status": "<string>",
      "status_langganan": "<string>",
      "catatan": "<string|null>",
      "id_sales": "<integer|null>",
      "nama_sales": "<string|null>",
      "id_router": "<string|null>",
      "jns_koneksi": "<string|null>",
      "tgl_jatuh_tempo": "<string|null>",
      "prorate": "<string|null>",
      "jns_blokir": "<string|null>",
      "latitude": "<string|null>",
      "longitude": "<string|null>",
      "modul_olt": "<string|null>",
      "olt_port": "<string|null>",
      "index_onu": "<string|null>",
      "sn_onu": "<string|null>",
      "insentif_sales": "<string|null>",
      "metode_insentif": "<string|null>",
      "nominal_insentif": "<string|null>",
      "id_olt": "<null>",
      "id_odc": "<null>",
      "id_odp": "<null>",
      "nama_olt": "<string|null>",
      "nama_odc": "<null>",
      "nama_odp": "<null>",
      "status_koneksi": "<string|null>",
      "waktu_pengecekan": "<datetime|null>",
      "jns_tagihan": "<string|null>",
      "tgl_tagihan": "<date|null>",
      "tgl_isolir": "<date|null>"
    }
  ],
  "links": {
    "first": "<string|null>",
    "last": "<string|null>",
    "prev": "<string|null>",
    "next": "<string|null>"
  },
  "meta": {
    "current_page": "<integer>",
    "from": "<integer|null>",
    "last_page": "<integer>",
    "links": [
      {
        "url": "<string|null>",
        "label": "<string>",
        "page": "<integer|null>",
        "active": "<boolean>"
      }
    ],
    "path": "<string>",
    "per_page": "<integer>",
    "to": "<integer|null>",
    "total": "<integer>"
  }
}
```

### `GET /tipe-pembayaran`

```json
{
  "data": [
    {
      "id_tipe_pembayaran": "<integer>",
      "account": "<integer>",
      "nama_tipe": "<string>",
      "keterangan": "<string>",
      "harga": "<integer>",
      "jns_tipe_pembayaran": "<string>",
      "id_lokasi": "<integer>",
      "profile": "<string|null>",
      "registrasi": "<string>",
      "filter_lokasi": "<string>",
      "fee": "<null>",
      "wkt_update": "<null>",
      "id_area": "<null>",
      "sub_area": "<null>",
      "nama_area": "<null>",
      "nama_cabang": "<null>",
      "id_cabang": "<null>"
    }
  ],
  "links": {
    "first": "<string|null>",
    "last": "<string|null>",
    "prev": "<string|null>",
    "next": "<string|null>"
  },
  "meta": {
    "current_page": "<integer>",
    "from": "<integer|null>",
    "last_page": "<integer>",
    "links": [
      {
        "url": "<string|null>",
        "label": "<string>",
        "page": "<integer|null>",
        "active": "<boolean>"
      }
    ],
    "path": "<string>",
    "per_page": "<integer>",
    "to": "<integer|null>",
    "total": "<integer>"
  }
}
```
