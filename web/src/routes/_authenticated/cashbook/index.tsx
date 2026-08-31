import { createFileRoute, redirect } from '@tanstack/react-router'
import { Cashbook } from '@/features/cashbook'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/cashbook/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'cashbook:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Cashbook,
})
