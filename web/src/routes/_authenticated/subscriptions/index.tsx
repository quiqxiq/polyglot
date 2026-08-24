import { createFileRoute, redirect } from '@tanstack/react-router'
import { Subscriptions } from '@/features/subscriptions'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/subscriptions/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'billing:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Subscriptions,
})
