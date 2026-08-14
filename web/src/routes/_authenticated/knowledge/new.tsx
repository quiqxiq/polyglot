import { createFileRoute, redirect } from '@tanstack/react-router'
import { KnowledgeEditorPage } from '@/features/knowledge/editor'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/knowledge/new')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'knowledge:write')) {
      throw redirect({ to: '/403' })
    }
  },
  component: () => <KnowledgeEditorPage mode='create' />,
})
