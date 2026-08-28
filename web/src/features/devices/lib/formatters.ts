/**
 * Utility functions for formatting bandwidth, timestamps, and error messages
 */

export function formatBps(bps: number): string {
  if (!bps || bps <= 0) return '0 bps'
  const units = ['bps', 'Kbps', 'Mbps', 'Gbps', 'Tbps']
  let v = bps
  let i = 0
  while (v >= 1000 && i < units.length - 1) {
    v /= 1000
    i++
  }
  return `${v.toFixed(v >= 100 || i === 0 ? 0 : 1)} ${units[i]}`
}

export function formatDateInput(d: Date): string {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

export function formatTimeInput(d: Date): string {
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

export function getErrorMessage(err: unknown, fallback = 'An unexpected error occurred'): string {
  if (!err) return fallback
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  if (typeof err === 'object' && 'rawMessage' in err && typeof (err as { rawMessage: unknown }).rawMessage === 'string') {
    return (err as { rawMessage: string }).rawMessage
  }
  return fallback
}

