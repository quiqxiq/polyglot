import { createFileRoute } from '@tanstack/react-router'
import { KnowledgeEditorPage } from '@/features/knowledge/editor'

export const Route = createFileRoute('/_authenticated/knowledge/new')({
  component: () => <KnowledgeEditorPage mode='create' />,
})
