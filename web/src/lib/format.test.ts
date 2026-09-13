import { describe, expect, it } from 'vitest'
import {
  formatBandwidthKbps,
  formatCurrency,
  formatUnixDate,
  toWaPhone,
} from './format'

describe('formatCurrency', () => {
  it('formats IDR without decimals', () => {
    expect(formatCurrency(150000)).toBe('Rp\u00A0150.000')
  })

  it('handles zero, bigint and undefined', () => {
    expect(formatCurrency(0)).toBe('Rp\u00A00')
    expect(formatCurrency(250000n)).toBe('Rp\u00A0250.000')
    expect(formatCurrency(undefined)).toBe('Rp\u00A00')
  })
})

describe('formatUnixDate', () => {
  // Tengah hari UTC agar tanggal sama di berbagai timezone umum.
  const noonUtc = BigInt(Date.UTC(2025, 10, 22, 12) / 1000)

  it('formats unix seconds to Indonesian short date', () => {
    expect(formatUnixDate(noonUtc)).toBe('22 Nov 2025')
  })

  it('formats long month style', () => {
    expect(formatUnixDate(noonUtc, 'long')).toBe('22 November 2025')
  })

  it('returns dash for empty and invalid values', () => {
    expect(formatUnixDate(0)).toBe('-')
    expect(formatUnixDate(undefined)).toBe('-')
  })
})

describe('formatBandwidthKbps', () => {
  it('formats kbps and mbps sides', () => {
    expect(formatBandwidthKbps(10000, 5000)).toBe('10M/5M')
    expect(formatBandwidthKbps(512, 256)).toBe('512k/256k')
  })

  it('handles partial and empty input', () => {
    expect(formatBandwidthKbps(10000, 0)).toBe('10M/0')
    expect(formatBandwidthKbps()).toBe('')
  })
})

describe('toWaPhone', () => {
  it('normalizes local numbers to 62 prefix', () => {
    expect(toWaPhone('081234567890')).toBe('6281234567890')
  })

  it('keeps international format and strips separators', () => {
    expect(toWaPhone('62 812-3456-7890')).toBe('6281234567890')
  })

  it('returns empty string for empty input', () => {
    expect(toWaPhone('')).toBe('')
    expect(toWaPhone(undefined)).toBe('')
  })
})
