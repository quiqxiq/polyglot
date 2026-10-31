import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { OltListPage } from '@/features/olt'

export const Route = createFileRoute('/_authenticated/olt/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'olt:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: OltListPage,
})
