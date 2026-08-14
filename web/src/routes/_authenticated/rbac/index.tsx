import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { RBAC } from '@/features/rbac'

export const Route = createFileRoute('/_authenticated/rbac/')({
  // Owner-only: rbac:manage adalah satu-satunya permission untuk halaman ini
  // dan hanya role owner yang memilikinya. Non-owner di-redirect ke 403.
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'rbac:manage')) {
      throw redirect({ to: '/403' })
    }
  },
  component: RBAC,
})
