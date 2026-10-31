import { createFileRoute } from '@tanstack/react-router'
import { SettingsPaymentMethods } from '@/features/settings/payment-methods'

export const Route = createFileRoute('/_authenticated/settings/payment-methods')({
  component: SettingsPaymentMethods,
})
