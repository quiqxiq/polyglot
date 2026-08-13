import { createFileRoute } from '@tanstack/react-router'
import { WhatsAppDevices } from '@/features/whatsapp'

export const Route = createFileRoute('/_authenticated/whatsapp/')({
  component: WhatsAppDevices,
})
