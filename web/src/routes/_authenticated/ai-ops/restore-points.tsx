import { createFileRoute, redirect } from '@tanstack/react-router'
import { RestorePointsPage } from '@/features/ai-ops/components/restore-points-page'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/ai-ops/restore-points')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'restore:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: RestorePointsPage,
})
