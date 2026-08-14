// Katalog resource & action yang dikenal sistem — dipakai untuk picker
// tambah policy. Harus sinkron dengan internal/adapter/auth/procedure_permissions.go.
export const RBAC_RESOURCES = [
  'device',
  'customer',
  'conversation',
  'knowledge',
  'billing',
  'whatsapp',
  'user',
  'rbac',
  'llmconfig',
  'technician',
  'probe',
  'hotspot',
] as const

export const RBAC_ACTIONS = [
  'read',
  'write',
  'manage',
  'command',
  'embed',
  'message',
] as const

export const ALL = '*'

// composeObject merangkai objek policy "resource:action" dari pilihan UI.
// - resource = ALL      -> ".*" (semua resource)
// - action   = ALL      -> "resource:.*" (semua aksi pada resource itu)
// - keduanya spesifik    -> "resource:action"
export function composeObject(resource: string, action: string): string {
  if (resource === ALL) return '.*'
  if (action === ALL) return `${resource}:.*`
  return `${resource}:${action}`
}

// resourceLabel memisahkan kembali objek policy menjadi (resource, action)
// untuk tampilan tabel.
export function splitObject(obj: string): { resource: string; action: string } {
  if (obj === '.*') return { resource: ALL, action: ALL }
  const idx = obj.indexOf(':')
  if (idx === -1) return { resource: obj, action: '' }
  const action = obj.slice(idx + 1)
  return { resource: obj.slice(0, idx), action: action === '*' ? ALL : action }
}
