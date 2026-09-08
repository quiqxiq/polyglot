/**
 * Memformat string username/identitas teknis (misal: "susanto_madek" atau "susanto-madek")
 * menjadi format nama manusia Title Case ("Susanto Madek").
 * Jika string tidak mengandung delimiter '_' atau '-', nilai asli dipertahankan seperti biasa.
 */
export function formatAutoName(raw: string): string {
  const s = raw.trim()
  if (!s) return ''
  const hasDelimiter = s.includes('_') || s.includes('-')
  if (!hasDelimiter) {
    return s
  }

  return s
    .replace(/[_-]+/g, ' ')
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

/**
 * Memecah pola konvensi "nama_alamat" atau "nama-alamat" menjadi bagian Nama dan Alamat.
 * Contoh: "susanto_madek" -> { name: "Susanto", address: "Madek" }
 * Contoh: "budi_santoso_rt01" -> { name: "Budi Santoso", address: "Rt01" }
 * Jika tidak ada delimiter atau hanya 1 bagian, address bernilai undefined.
 */
export function parseUsernameNameAndAddress(raw: string): {
  name: string
  address?: string
} {
  const s = raw.trim()
  if (!s) return { name: '' }
  const delimiter = s.includes('_') ? '_' : s.includes('-') ? '-' : null
  if (!delimiter) {
    return { name: s }
  }

  const parts = s.split(delimiter).filter(Boolean)
  if (parts.length >= 2) {
    // Bagian terakhir adalah alamat/lokasi (misal "madek" atau "rt01"), bagian depan adalah nama
    const nameParts = parts.slice(0, parts.length - 1)
    const addressPart = parts[parts.length - 1]

    const formattedName = nameParts
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
      .join(' ')
    const formattedAddress =
      addressPart.charAt(0).toUpperCase() + addressPart.slice(1).toLowerCase()

    return {
      name: formattedName,
      address: formattedAddress,
    }
  }

  return { name: formatAutoName(raw) }
}
