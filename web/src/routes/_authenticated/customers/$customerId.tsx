import { createFileRoute, redirect } from '@tanstack/react-router'
import { CustomerDetailPage } from '@/features/customer/components/detail/customer-detail-page'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/customers/$customerId')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'customer:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: CustomerDetailPage,
})
