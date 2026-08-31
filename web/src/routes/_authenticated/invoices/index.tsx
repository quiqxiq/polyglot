import { createFileRoute, redirect } from '@tanstack/react-router'
import { Invoices } from '@/features/invoices'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/invoices/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'billing:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Invoices,
})
