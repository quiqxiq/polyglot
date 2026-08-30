export type ServiceType = 'PPPOE' | 'HOTSPOT' | 'DEDICATED'

// TYPE_ONLY_FIELDS: field yang HANYA relevan untuk tipe tertentu.
// Semua field lain yang tidak ada di ANY_TYPE_ONLY (nama, harga bulanan, bandwidth dasar, instalasi, pajak, deskripsi)
// bersifat umum dan selalu tampil.
const TYPE_ONLY_FIELDS: Record<ServiceType, ReadonlySet<string>> = {
  // PPPoE langganan bulanan: pool IP pelanggan, address-list, parent queue, burst.
  // Tidak ada konsep voucher/Mikhmon (validity, expireMode, sharedUsers, sellingPrice, lockUser).
  PPPOE: new Set([
    'parentQueue',
    'addressList',
    'remoteAddressPool',
    'simultaneousUse',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
  // Hotspot Mikhmon/Voucher: harga jual reseller, validity/expire, shared users, lock MAC/server, IP pool.
  HOTSPOT: new Set([
    'sellingPrice',
    'validity',
    'validityMode',
    'expireMode',
    'sharedUsers',
    'ipPoolName',
    'parentQueue',
    'lockUser',
    'lockServer',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
  // Dedicated CIR: queue + routing saja; tanpa validity/voucher/selling price.
  DEDICATED: new Set([
    'parentQueue',
    'addressList',
    'remoteAddressPool',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
}

// Field di luar gabungan set type-only adalah field umum dan wajib selalu tampil.
const ANY_TYPE_ONLY = new Set(
  Object.values(TYPE_ONLY_FIELDS).flatMap((set) => [...set]),
)

export function isFieldVisible(field: string, serviceType: string): boolean {
  if (!ANY_TYPE_ONLY.has(field)) return true // field umum (nama, price, bandwidth, dll) → selalu tampil
  const set = TYPE_ONLY_FIELDS[serviceType as ServiceType]
  if (!set) return true // tipe tak dikenal → fallback aman
  return set.has(field)
}

// HIDDEN_FOR_TYPE: pengecualian eksplisit bila ada
const HIDDEN_FOR_TYPE: Record<ServiceType, ReadonlySet<string>> = {
  PPPOE: new Set(['expireMode', 'validity', 'validityMode', 'sellingPrice', 'sharedUsers', 'lockUser', 'lockServer']),
  HOTSPOT: new Set(['remoteAddressPool', 'addressList']),
  DEDICATED: new Set(['expireMode', 'validity', 'validityMode', 'sellingPrice', 'sharedUsers', 'lockUser', 'lockServer']),
}

export function isFieldHidden(field: string, serviceType: string): boolean {
  const set = HIDDEN_FOR_TYPE[serviceType as ServiceType]
  return set ? set.has(field) : false
}
