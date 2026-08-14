import { createFileRoute, redirect } from '@tanstack/react-router'
import { Knowledge } from '@/features/knowledge'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/knowledge/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'knowledge:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Knowledge,
})
