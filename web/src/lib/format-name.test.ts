import { describe, expect, it } from 'vitest'
import { formatAutoName, parseUsernameNameAndAddress } from './format-name'

describe('formatAutoName', () => {
  it('formats underscore delimited usernames to Title Case', () => {
    expect(formatAutoName('susanto_madek')).toBe('Susanto Madek')
    expect(formatAutoName('budi_santoso_rt01')).toBe('Budi Santoso Rt01')
  })

  it('formats hyphen delimited usernames to Title Case', () => {
    expect(formatAutoName('susanto-madek')).toBe('Susanto Madek')
    expect(formatAutoName('paket-home-10m')).toBe('Paket Home 10m')
  })

  it('preserves names without delimiters as is', () => {
    expect(formatAutoName('susanto')).toBe('susanto')
    expect(formatAutoName('JohnDoe')).toBe('JohnDoe')
  })

  it('handles empty and whitespace-only strings gracefully', () => {
    expect(formatAutoName('')).toBe('')
    expect(formatAutoName('   ')).toBe('')
  })
})

describe('parseUsernameNameAndAddress', () => {
  it('splits underscore format nama_alamat into name and address', () => {
    const res = parseUsernameNameAndAddress('susanto_madek')
    expect(res.name).toBe('Susanto')
    expect(res.address).toBe('Madek')
  })

  it('splits multi-part names with address as last segment', () => {
    const res = parseUsernameNameAndAddress('budi_santoso_rt01')
    expect(res.name).toBe('Budi Santoso')
    expect(res.address).toBe('Rt01')
  })

  it('splits hyphen format nama-alamat', () => {
    const res = parseUsernameNameAndAddress('susanto-madek')
    expect(res.name).toBe('Susanto')
    expect(res.address).toBe('Madek')
  })

  it('returns only name when no delimiter is present', () => {
    const res = parseUsernameNameAndAddress('susanto')
    expect(res.name).toBe('susanto')
    expect(res.address).toBeUndefined()
  })
})
