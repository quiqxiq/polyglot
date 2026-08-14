import { createFileRoute } from '@tanstack/react-router'
import { KnowledgeEditorPage } from '@/features/knowledge/editor'

export const Route = createFileRoute('/_authenticated/knowledge/$id/edit')({
  component: RouteComponent,
})

// eslint-disable-next-line react-refresh/only-export-components
function RouteComponent() {
  const { id } = Route.useParams()
  return <KnowledgeEditorPage mode='edit' id={id} />
}
