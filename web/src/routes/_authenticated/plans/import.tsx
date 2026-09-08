import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { PlansImportPage } from '@/features/billing/components/plans-import-page'

export const Route = createFileRoute('/_authenticated/plans/import')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'billing:manage')) {
      throw redirect({ to: '/403' })
    }
  },
  component: PlansImportPage,
})
