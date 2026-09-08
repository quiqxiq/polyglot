import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { CustomersImportPage } from '@/features/customer/components/customers-import-page'

export const Route = createFileRoute('/_authenticated/customers/import')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'customer:manage')) {
      throw redirect({ to: '/403' })
    }
  },
  component: CustomersImportPage,
})
