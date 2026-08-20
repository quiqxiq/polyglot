import { useAuthStore } from '@/stores/auth-store'

// canPermission replicates backend Casbin enforcement for a permission string
// like "knowledge:read" against the flat effective-permission list the server
// returns (GetImplicitPermissionsForUser, format "resource:action").
//
// Backend matcher (internal/config/rbac_model.conf):
//   regexMatch(r.obj, p.obj) && (p.act == "*" || p.act == r.act)
// where r.obj IS the full "resource:action" string (e.g. "knowledge:read").
// So each flat permission "pResource:pAction" matches when pResource (a regex,
// e.g. "knowledge:.*" or ".*") matches the full request object AND pAction is
// "*" or equals the request action.
export function canPermission(
  permissions: string[] | undefined,
  permission: string,
): boolean {
  if (!permissions || permissions.length === 0) return false

  const actionIdx = permission.lastIndexOf(':')
  if (actionIdx <= 0) return false
  const requestAction = permission.slice(actionIdx + 1)

  return permissions.some((flat) => {
    const pIdx = flat.lastIndexOf(':')
    if (pIdx <= 0) return false
    const pResource = flat.slice(0, pIdx)
    const pAction = flat.slice(pIdx + 1)
    if (pAction !== '*' && pAction !== requestAction) return false
    try {
      return new RegExp(pResource).test(permission)
    } catch {
      return pResource === permission
    }
  })
}

// useCan returns whether the signed-in user holds the given permission
// (e.g. useCan('user:read')). Re-renders reactively when user changes.
export function useCan(permission: string): boolean {
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  return canPermission(permissions, permission)
}
