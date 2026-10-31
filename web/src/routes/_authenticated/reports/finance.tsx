import { createFileRoute, redirect } from '@tanstack/react-router'
import { FinanceReportPage } from '@/features/reports/components/finance-report-page'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/reports/finance')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'report:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: FinanceReportPage,
})
