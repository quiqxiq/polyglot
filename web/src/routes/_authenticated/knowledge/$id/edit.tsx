import { createFileRoute, redirect } from '@tanstack/react-router'
import { KnowledgeEditorPage } from '@/features/knowledge/editor'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'

export const Route = createFileRoute('/_authenticated/knowledge/$id/edit')({
  beforeLoad: () => {
    const permissions = useAuthStore.getState().auth.user?.permissions
    if (!canPermission(permissions, 'knowledge:write')) {
      throw redirect({ to: '/403' })
    }
  },
  component: RouteComponent,
})

// eslint-disable-next-line react-refresh/only-export-components
function RouteComponent() {
  const { id } = Route.useParams()
  return <KnowledgeEditorPage mode='edit' id={id} />
}
