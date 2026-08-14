// Katalog role sistem — harus sinkron dengan KnownRoles di backend
// (internal/adapter/connect/auth/user_handler.go) dan policy seeder.
export const ROLE_META: Record<string, { label: string; className: string }> = {
  owner: {
    label: 'Owner',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400',
  },
  admin: {
    label: 'Admin',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400',
  },
  agent: {
    label: 'Agent',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  },
  teknisi: {
    label: 'Teknisi',
    className: 'bg-violet-500/15 text-violet-700 dark:text-violet-400',
  },
}

export const ROLE_OPTIONS = Object.entries(ROLE_META).map(([value, meta]) => ({
  value,
  label: meta.label,
}))

export function roleLabel(role: string): string {
  return ROLE_META[role]?.label ?? role
}

export function roleClassName(role: string): string {
  return ROLE_META[role]?.className ?? ''
}
