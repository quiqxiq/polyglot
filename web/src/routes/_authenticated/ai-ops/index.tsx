import { createFileRoute, redirect } from '@tanstack/react-router'
import { AIOpsChatPage } from '@/features/ai-ops/components/chat-page'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/ai-ops/')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'ai:ops')) {
      throw redirect({ to: '/403' })
    }
  },
  component: AIOpsChatPage,
})
