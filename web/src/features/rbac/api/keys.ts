export const rbacKeys = {
  all: ['rbac'] as const,
  policies: () => [...rbacKeys.all, 'policies'] as const,
  roleAssignments: () => [...rbacKeys.all, 'role-assignments'] as const,
}
