function toLocaleTag(language: string): string {
  return language === 'en' ? 'en-GB' : 'id-ID'
}

/** Format an RFC3339 timestamp returned by the OLT API. */
export function formatDateTime(
  value: string | undefined,
  language: string
): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString(toLocaleTag(language), {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatRxPower(value: number | undefined): string {
  if (value === undefined || value === null || Number.isNaN(value)) return '—'
  return `${value.toFixed(2)} dBm`
}

export function formatDistance(value: number | undefined): string {
  if (value === undefined || value === null) return '—'
  if (value >= 1000) return `${(value / 1000).toFixed(2)} km`
  return `${value} m`
}

export function formatTemperature(value: number | undefined): string {
  if (value === undefined || value === null) return '—'
  return `${value.toFixed(1)} °C`
}

export function formatLatency(value: bigint | number | undefined): string {
  if (value === undefined || value === null) return '—'
  return `${Number(value)} ms`
}

/** PON/ONU position shorthand used by confirm dialogs. */
export function positionLabel(ponIndex: number, onuIndex: number): string {
  return `PON ${ponIndex} / ONU ${onuIndex}`
}
