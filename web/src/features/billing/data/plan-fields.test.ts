import { describe, expect, it } from 'vitest'
import { isFieldHidden, isFieldVisible } from './plan-fields'

describe('isFieldVisible', () => {
  it('hides hotspot-only fields for PPPOE', () => {
    expect(isFieldVisible('sharedUsers', 'PPPOE')).toBe(false)
    expect(isFieldVisible('expireMode', 'PPPOE')).toBe(false)
    expect(isFieldVisible('ipPoolName', 'PPPOE')).toBe(false)
    expect(isFieldVisible('lockUser', 'PPPOE')).toBe(false)
    expect(isFieldVisible('lockServer', 'PPPOE')).toBe(false)
  })
  it('shows ppp fields for PPPOE', () => {
    expect(isFieldVisible('validity', 'PPPOE')).toBe(true)
    expect(isFieldVisible('validityMode', 'PPPOE')).toBe(true)
    expect(isFieldVisible('simultaneousUse', 'PPPOE')).toBe(true)
    expect(isFieldVisible('parentQueue', 'PPPOE')).toBe(true)
    expect(isFieldVisible('addressList', 'PPPOE')).toBe(true)
    expect(isFieldVisible('remoteAddressPool', 'PPPOE')).toBe(true)
    expect(isFieldVisible('burstDownloadKbps', 'PPPOE')).toBe(true)
  })
  it('hotspot shows mikhmon fields, hides addressList & simultaneousUse', () => {
    expect(isFieldVisible('expireMode', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('sharedUsers', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('ipPoolName', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('lockUser', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('lockServer', 'HOTSPOT')).toBe(true)
    expect(isFieldVisible('addressList', 'HOTSPOT')).toBe(false)
    expect(isFieldVisible('simultaneousUse', 'HOTSPOT')).toBe(false)
    expect(isFieldVisible('validity', 'HOTSPOT')).toBe(true)
  })
  it('dedicated shows cir fields only', () => {
    expect(isFieldVisible('parentQueue', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('addressList', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('remoteAddressPool', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('burstDownloadKbps', 'DEDICATED')).toBe(true)
    expect(isFieldVisible('sharedUsers', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('validity', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('validityMode', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('expireMode', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('ipPoolName', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('lockUser', 'DEDICATED')).toBe(false)
    expect(isFieldVisible('simultaneousUse', 'DEDICATED')).toBe(false)
  })
  it('remote address pool hidden for HOTSPOT', () => {
    expect(isFieldVisible('remoteAddressPool', 'HOTSPOT')).toBe(false)
  })
  it('common fields always visible for all types', () => {
    const common = ['name','serviceType','bandwidthDownloadKbps','bandwidthUploadKbps','price','sellingPrice','installationFee','taxPercent','isActive','description'] as const
    const types = ['PPPOE','HOTSPOT','DEDICATED'] as const
    for (const f of common)
      for (const t of types)
        expect(isFieldVisible(f, t), `${f}@${t}`).toBe(true)
  })
  it('unknown type falls back to visible', () => {
    expect(isFieldVisible('anything', 'MYSTERY')).toBe(true)
  })
})

describe('isFieldHidden', () => {
  it('hides expireMode for PPPOE only', () => {
    expect(isFieldHidden('expireMode', 'PPPOE')).toBe(true)
    expect(isFieldHidden('expireMode', 'HOTSPOT')).toBe(false)
    expect(isFieldHidden('expireMode', 'DEDICATED')).toBe(true)
  })
  it('other fields unaffected', () => {
    expect(isFieldHidden('validity', 'PPPOE')).toBe(false)
    expect(isFieldHidden('sharedUsers', 'HOTSPOT')).toBe(false)
  })
})
