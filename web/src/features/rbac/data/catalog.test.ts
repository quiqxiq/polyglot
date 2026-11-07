import { describe, expect, it } from 'vitest'
import {
  ALL_PERMISSION_IDS,
  RBAC_MODULE_GROUPS,
  RBAC_RESOURCES,
  composeObject,
  splitObject,
} from './catalog'

describe('RBAC catalog & module groups', () => {
  it('contains the dedicated Web Dial module group', () => {
    const dialGroup = RBAC_MODULE_GROUPS.find((g) => g.id === 'dial')
    expect(dialGroup).toBeDefined()
    expect(dialGroup?.label).toContain('Web Dial')
    expect(dialGroup?.permissions.length).toBe(8)
  })

  it('contains all 8 dial page permissions with correct prefixes and non-empty metadata', () => {
    const dialGroup = RBAC_MODULE_GROUPS.find((g) => g.id === 'dial')!
    const dialPermIds = dialGroup.permissions.map((p) => p.id)

    const expectedDialPerms = [
      'dial:dashboard:read',
      'dial:ppp-active:read',
      'dial:ppp-non-active:read',
      'dial:ppp-secrets:read',
      'dial:ppp-profiles:read',
      'dial:logs:read',
      'dial:monitor:read',
      'dial:settings:read',
    ]

    expect(dialPermIds).toEqual(expectedDialPerms)

    for (const perm of dialGroup.permissions) {
      expect(perm.id.startsWith('dial:')).toBe(true)
      expect(perm.resource).toBe('dial')
      expect(perm.action.length).toBeGreaterThan(0)
      expect(perm.label.length).toBeGreaterThan(0)
      expect(perm.description.length).toBeGreaterThan(0)
    }
  })

  it('includes ppp:inactive:read in the ppp module group', () => {
    const pppGroup = RBAC_MODULE_GROUPS.find((g) => g.id === 'ppp')!
    const pppPermIds = pppGroup.permissions.map((p) => p.id)
    expect(pppPermIds).toContain('ppp:inactive:read')
  })

  it('includes all dial permissions in ALL_PERMISSION_IDS', () => {
    expect(ALL_PERMISSION_IDS).toContain('dial:dashboard:read')
    expect(ALL_PERMISSION_IDS).toContain('dial:ppp-active:read')
    expect(ALL_PERMISSION_IDS).toContain('dial:settings:read')
    expect(ALL_PERMISSION_IDS).toContain('ppp:inactive:read')
  })

  it('includes dial in RBAC_RESOURCES', () => {
    expect(RBAC_RESOURCES).toContain('dial')
  })

  it('correctly splits and composes dial permissions', () => {
    const split = splitObject('dial:ppp-active:read')
    expect(split.resource).toBe('dial')
    expect(split.action).toBe('ppp-active:read')

    const composed = composeObject(split.resource, split.action)
    expect(composed).toBe('dial:ppp-active:read')
  })
})
