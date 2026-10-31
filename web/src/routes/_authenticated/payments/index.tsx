import { createFileRoute, redirect } from '@tanstack/react-router'
import { Payments } from '@/features/payments'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/payments/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'billing:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Payments,
})
