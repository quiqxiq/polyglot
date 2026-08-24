export type ServiceType = 'PPPOE' | 'HOTSPOT' | 'DEDICATED'

// TYPE_ONLY_FIELDS: field yang HANYA relevan untuk tipe tertentu.
// Semua field lain (nama, bandwidth, harga, dst.) bersifat umum dan selalu tampil.
const TYPE_ONLY_FIELDS: Record<ServiceType, ReadonlySet<string>> = {
  // PPPoE langganan bulanan: validity + ppp routing; tanpa konsep voucher Mikhmon.
  PPPOE: new Set([
    'validity',
    'validityMode',
    'simultaneousUse',
    'parentQueue',
    'addressList',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
  // Hotspot Mikhmon: validity/expire + shared users + lock; tanpa ppp routing.
  HOTSPOT: new Set([
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
  // Dedicated CIR: queue + routing saja; tanpa validity/voucher.
  DEDICATED: new Set([
    'parentQueue',
    'addressList',
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ]),
}

// DEVIASI dari snippet rencana: `set.has(field)` saja menyembunyikan field umum
// (mis. `name`) karena tidak ada di set manapun. Field di luar gabungan set
// type-only adalah field umum dan wajib selalu tampil.
const ANY_TYPE_ONLY = new Set(
  Object.values(TYPE_ONLY_FIELDS).flatMap((set) => [...set]),
)

export function isFieldVisible(field: string, serviceType: string): boolean {
  if (!ANY_TYPE_ONLY.has(field)) return true // field umum → selalu tampil
  const set = TYPE_ONLY_FIELDS[serviceType as ServiceType]
  if (!set) return true // tipe tak dikenal → fallback aman: tampilkan semua
  return set.has(field)
}
