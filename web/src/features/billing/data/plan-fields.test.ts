import { describe, expect, it } from 'vitest'
import { isFieldHidden, isFieldVisible } from './plan-fields'

describe('isFieldVisible', () => {
  it('hides hotspot-only fields for PPPOE', () => {
    expect(isFieldVisible('sharedUsers', 'PPPOE')).toBe(false)
    expect(isFieldVisible('ipPoolName', 'PPPOE')).toBe(false)
  })
  it('shows ppp fields for PPPOE', () => {
    expect(isFieldVisible('parentQueue', 'PPPOE')).toBe(true)
    expect(isFieldVisible('addressList', 'PPPOE')).toBe(true)
    expect(isFieldVisible('remoteAddressPool', 'PPPOE')).toBe(true)
    expect(isFieldVisible('burstDownloadKbps', 'PPPOE')).toBe(true)
  })
  it('hotspot shows sharedUsers and ipPoolName, hides remoteAddressPool', () => {
    expect(isFieldVisible('sharedUsers', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('ipPoolName', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('addressList', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('remoteAddressPool', 'HOTSPOT')).toBe(false)
  })
  it('dedicated shows cir fields only', () => {
    expect(isFieldVisible('parentQueue', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('addressList', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('remoteAddressPool', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('burstDownloadKbps', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('sharedUsers', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('ipPoolName', 'DEDICATED')).toBe(false)
  })
  it('common fields always visible for all types', () => {
    const common = [
      'name',
      'serviceType',
      'bandwidthDownloadKbps',
      'bandwidthUploadKbps',
      'price',
      'installationFee',
      'taxPercent',
      'isActive',
      'description',
    ] as const
    const types = ['PPPOE', 'HOTSPOT', 'DEDICATED'] as const
    for (const f of common)
      for (const t of types)
        expect(isFieldVisible(f, t), `${f}@${t}`).toBe(true)
  })
  it('unknown type falls back to visible', () => {
    expect(isFieldVisible('anything', 'MYSTERY')).toBe(true)
  })
})

describe('isFieldHidden', () => {
  it('hides sharedUsers for PPPOE & DEDICATED', () => {
    expect(isFieldHidden('sharedUsers', 'PPPOE')).toBe(true)
    expect(isFieldHidden('sharedUsers', 'HOTSPOT')).toBe(false)
    expect(isFieldHidden('sharedUsers', 'DEDICATED')).toBe(true)
  })
  it('hides remoteAddressPool for HOTSPOT', () => {
    expect(isFieldHidden('remoteAddressPool', 'HOTSPOT')).toBe(true)
    expect(isFieldHidden('remoteAddressPool', 'PPPOE')).toBe(false)
  })
})
