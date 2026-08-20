export const rbacKeys = {
  all: ['rbac'] as const,
  policies: () => [...rbacKeys.all, 'policies'] as const,
  assignments: () => [...rbacKeys.all, 'assignments'] as const,
}
