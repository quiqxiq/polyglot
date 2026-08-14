import { describe, expect, it } from 'vitest'
import { canPermission } from './use-can'

// Flat effective permissions come from the backend as "resource:action" —
// the resource part is a regex (e.g. "knowledge:.*"), action usually "*".
// owner: full access; admin: everything except rbac; agent: conversations etc.
const ownerPerms = ['.*:*']
const adminPerms = [
  'device:.*:*',
  'knowledge:.*:*',
  'user:.*:*',
  'conversation:.*:*',
  'whatsapp:.*:*',
]
const agentPerms = [
  'conversation:.*:*',
  'customer:read:*',
  'knowledge:read:*',
  'billing:read:*',
  'whatsapp:read:*',
]
const teknisiPerms = [
  'device:read:*',
  'device:command:*',
  'knowledge:read:*',
  'technician:read:*',
  'probe:read:*',
  'hotspot:read:*',
]

describe('canPermission', () => {
  it('denies when there are no permissions', () => {
    expect(canPermission(undefined, 'user:read')).toBe(false)
    expect(canPermission([], 'user:read')).toBe(false)
  })

  it('owner wildcard matches everything', () => {
    expect(canPermission(ownerPerms, 'user:read')).toBe(true)
    expect(canPermission(ownerPerms, 'rbac:manage')).toBe(true)
    expect(canPermission(ownerPerms, 'knowledge:write')).toBe(true)
  })

  it('admin wildcard per resource matches every action of that resource', () => {
    expect(canPermission(adminPerms, 'user:read')).toBe(true)
    expect(canPermission(adminPerms, 'user:manage')).toBe(true)
    expect(canPermission(adminPerms, 'knowledge:write')).toBe(true)
  })

  it('admin has no rbac:manage (matches backend deny)', () => {
    expect(canPermission(adminPerms, 'rbac:manage')).toBe(false)
  })

  it('agent read-only: exact action from read policies only', () => {
    expect(canPermission(agentPerms, 'knowledge:read')).toBe(true)
    expect(canPermission(agentPerms, 'customer:read')).toBe(true)
    expect(canPermission(agentPerms, 'knowledge:write')).toBe(false)
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
    expect(canPermission(['knowledge:read:read'], 'knowledge:read')).toBe(true)
    expect(canPermission(['knowledge:read:read'], 'knowledge:write')).toBe(false)
  })
})
