const idrFormatter = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

export function formatCurrency(
  amount: number | bigint | null | undefined
): string {
  const num = Number(amount ?? 0)
  if (!Number.isFinite(num)) return idrFormatter.format(0)
  return idrFormatter.format(num)
}

export function formatUnixDate(
  value: bigint | number | string | null | undefined,
  style: 'short' | 'long' = 'short'
): string {
  const num = Number(value ?? 0)
  if (!num) return '-'
  const date = new Date(num * 1000)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: style === 'long' ? 'long' : 'short',
    year: 'numeric',
  })
}

export function formatBandwidthKbps(
  downloadKbps?: number,
  uploadKbps?: number
): string {
  const side = (kbps?: number) => {
    if (!kbps || kbps <= 0) return ''
    if (kbps >= 1000) {
      const mbps = kbps / 1000
      return `${Number.isInteger(mbps) ? mbps : mbps.toFixed(1)}M`
    }
    return `${kbps}k`
  }
  const dl = side(downloadKbps)
  const ul = side(uploadKbps)
  if (!dl && !ul) return ''
  return `${dl || '0'}/${ul || '0'}`
}

export function toWaPhone(phone: string | null | undefined): string {
  const digits = (phone ?? '').replace(/\D/g, '')
  if (!digits) return ''
  if (digits.startsWith('0')) return `62${digits.slice(1)}`
  return digits
}
