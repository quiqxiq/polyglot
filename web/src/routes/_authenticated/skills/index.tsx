import { createFileRoute } from '@tanstack/react-router'
import { Skills } from '@/features/skills'

export const Route = createFileRoute('/_authenticated/skills/')({
  component: Skills,
})
