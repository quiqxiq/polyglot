export type ServiceType = 'PPPOE' | 'HOTSPOT' | 'DEDICATED'

// TYPE_ONLY_FIELDS: field yang HANYA relevan untuk tipe tertentu.
// Semua field umum (nama, harga bulanan, bandwidth dasar, instalasi, pajak, deskripsi)
// tidak ada di ANY_TYPE_ONLY dan selalu tampil.
const TYPE_ONLY_FIELDS: Record<ServiceType, ReadonlySet<string>> = {
  // PPPoE langganan bulanan: pool IP pelanggan, address-list, parent queue, burst, timeout.
  PPPOE: new Set([
    'parentQueue',
    'addressList',
    'remoteAddressPool',
    'sessionTimeout',
    'idleTimeout',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
  // Hotspot Permanent: shared users (jumlah gadget login bersamaan), pool IP hotspot, address-list, parent queue, burst, timeout.
  HOTSPOT: new Set([
    'sharedUsers',
    'ipPoolName',
    'addressList',
    'parentQueue',
    'sessionTimeout',
    'idleTimeout',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
  // Dedicated CIR: queue + routing/pool + burst, timeout.
  DEDICATED: new Set([
    'parentQueue',
    'addressList',
    'remoteAddressPool',
    'sessionTimeout',
    'idleTimeout',
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
  PPPOE: new Set(['sharedUsers', 'ipPoolName']),
  HOTSPOT: new Set(['remoteAddressPool']),
  DEDICATED: new Set(['sharedUsers', 'ipPoolName']),
}

export function isFieldHidden(field: string, serviceType: string): boolean {
  const set = HIDDEN_FOR_TYPE[serviceType as ServiceType]
  return set ? set.has(field) : false
}
