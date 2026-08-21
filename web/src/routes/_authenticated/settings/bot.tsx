import { createFileRoute } from '@tanstack/react-router'
import { SettingsBot } from '@/features/settings/bot'

export const Route = createFileRoute('/_authenticated/settings/bot')({
  component: SettingsBot,
})
