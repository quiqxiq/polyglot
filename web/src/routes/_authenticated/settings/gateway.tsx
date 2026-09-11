import { createFileRoute } from '@tanstack/react-router'
import { SettingsGateway } from '@/features/settings/gateway'

export const Route = createFileRoute('/_authenticated/settings/gateway')({
  component: SettingsGateway,
})
