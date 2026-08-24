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
    'remoteAddressPool',
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
    'remoteAddressPool',
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

// HIDDEN_FOR_TYPE: pengecualian eksplisit — field disembunyikan untuk tipe ini
// meskipun secara umum tersedia. Keputusan owner: PPPOE & DEDICATED tidak
// memakai expire mode Mikhmon.
const HIDDEN_FOR_TYPE: Record<ServiceType, ReadonlySet<string>> = {
  PPPOE: new Set(['expireMode']),
  HOTSPOT: new Set(),
  DEDICATED: new Set(['expireMode']),
}

export function isFieldHidden(field: string, serviceType: string): boolean {
  const set = HIDDEN_FOR_TYPE[serviceType as ServiceType]
  return set ? set.has(field) : false // tipe tak dikenal → fallback aman: sembunyikan apa pun
}
