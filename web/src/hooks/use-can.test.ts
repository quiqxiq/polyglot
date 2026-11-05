import { describe, expect, it } from 'vitest'
import { canAllPermissions, canAnyPermission, canPermission } from './use-can'

// Flat effective permissions come from the backend as "resource:action" —
// the resource part is a regex (e.g. "knowledge:.*"), action usually "*".
// owner: full access; admin: everything except rbac; agent: conversations etc.
const ownerPerms = ['.*:*']
const adminPerms = [
  'device:.*:*',
  'skill:.*:*',
  'user:.*:*',
  'conversation:.*:*',
  'whatsapp:.*:*',
]
const agentPerms = [
  'conversation:.*:*',
  'customer:read:*',
  'skill:read:*',
  'billing:read:*',
  'whatsapp:read:*',
]
const teknisiPerms = [
  'device:read:*',
  'device:command:*',
  'skill:read:*',
  'technician:read:*',
  'probe:read:*',
  'hotspot:read:*',
  'dial:ppp-active:read:*',
]

describe('canPermission', () => {
  it('denies when there are no permissions', () => {
    expect(canPermission(undefined, 'user:read')).toBe(false)
    expect(canPermission([], 'user:read')).toBe(false)
  })

  it('owner wildcard matches everything', () => {
    expect(canPermission(ownerPerms, 'user:read')).toBe(true)
    expect(canPermission(ownerPerms, 'rbac:manage')).toBe(true)
    expect(canPermission(ownerPerms, 'skill:manage')).toBe(true)
  })

  it('admin wildcard per resource matches every action of that resource', () => {
    expect(canPermission(adminPerms, 'user:read')).toBe(true)
    expect(canPermission(adminPerms, 'user:manage')).toBe(true)
    expect(canPermission(adminPerms, 'skill:manage')).toBe(true)
  })

  it('admin has no rbac:manage (matches backend deny)', () => {
    expect(canPermission(adminPerms, 'rbac:manage')).toBe(false)
  })

  it('agent read-only: exact action from read policies only', () => {
    expect(canPermission(agentPerms, 'skill:read')).toBe(true)
    expect(canPermission(agentPerms, 'customer:read')).toBe(true)
    expect(canPermission(agentPerms, 'skill:manage')).toBe(false)
    expect(canPermission(agentPerms, 'user:read')).toBe(false)
    expect(canPermission(agentPerms, 'rbac:manage')).toBe(false)
  })

  it('teknisi command vs manage distinction', () => {
    expect(canPermission(teknisiPerms, 'device:read')).toBe(true)
    expect(canPermission(teknisiPerms, 'device:command')).toBe(true)
    expect(canPermission(teknisiPerms, 'device:manage')).toBe(false)
  })

  it('non-* actions are compared exactly', () => {
    // policy "customer:read:*" — resource regex must match full object
    expect(canPermission(['customer:read:*'], 'customer:read')).toBe(true)
    expect(canPermission(['customer:read:*'], 'customer:write')).toBe(false)
    // a non-wildcard action only matches itself
    expect(canPermission(['skill:read:read'], 'skill:read')).toBe(true)
    expect(canPermission(['skill:read:read'], 'skill:manage')).toBe(false)
  })

  it('anchored regex prevents un-anchored partial substring matching', () => {
    // "profile:.*:*" should NOT match "ppp:profile:read"
    expect(canPermission(['profile:.*:*'], 'profile:read')).toBe(true)
    expect(canPermission(['profile:.*:*'], 'ppp:profile:read')).toBe(false)

    // "dial:.*:*" should NOT match "ppp:active:read"
    expect(canPermission(['dial:.*:*'], 'dial:ppp-active:read')).toBe(true)
    expect(canPermission(['dial:.*:*'], 'ppp:active:read')).toBe(false)

    // Having dial permission does NOT grant polyglot permission
    expect(canPermission(['dial:ppp-active:read:*'], 'ppp:active:read')).toBe(false)
    expect(canPermission(['ppp:active:read:*'], 'dial:ppp-active:read')).toBe(false)
  })
})

describe('canAnyPermission', () => {
  it('returns true when required list is empty', () => {
    expect(canAnyPermission(agentPerms, [])).toBe(true)
    expect(canAnyPermission(undefined, [])).toBe(true)
  })

  it('returns true if any permission matches', () => {
    expect(canAnyPermission(teknisiPerms, ['user:read', 'device:read'])).toBe(true)
    expect(canAnyPermission(teknisiPerms, ['ppp:active:read', 'dial:ppp-active:read'])).toBe(true)
  })

  it('returns false if none match', () => {
    expect(canAnyPermission(teknisiPerms, ['user:read', 'user:manage'])).toBe(false)
  })
})

describe('canAllPermissions', () => {
  it('returns true when required list is empty', () => {
    expect(canAllPermissions(agentPerms, [])).toBe(true)
    expect(canAllPermissions(undefined, [])).toBe(true)
  })

  it('returns true only if all permissions match', () => {
    expect(canAllPermissions(teknisiPerms, ['device:read', 'device:command'])).toBe(true)
    expect(canAllPermissions(teknisiPerms, ['device:read', 'device:manage'])).toBe(false)
  })
})
