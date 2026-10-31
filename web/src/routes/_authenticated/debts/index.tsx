import { createFileRoute, redirect } from '@tanstack/react-router'
import { DebtsPage } from '@/features/debts'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/debts/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'debt:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: DebtsPage,
})
