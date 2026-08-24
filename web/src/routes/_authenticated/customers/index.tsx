import { createFileRoute, redirect } from '@tanstack/react-router'
import { Customers } from '@/features/customer'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/customers/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'customer:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Customers,
})
