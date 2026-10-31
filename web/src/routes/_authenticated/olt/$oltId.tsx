import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { OltDetailPage } from '@/features/olt/components/detail/olt-detail-page'

export const Route = createFileRoute('/_authenticated/olt/$oltId')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'olt:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: OltDetailPage,
})
