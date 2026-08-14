import { createFileRoute, redirect } from '@tanstack/react-router'
import { Chats } from '@/features/chats'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/chats/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'conversation:read')) {
      throw redirect({ to: '/403' })
    }
  },
  component: Chats,
})
