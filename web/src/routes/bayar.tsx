import { createFileRoute } from '@tanstack/react-router'
import { PayManualPage } from '@/features/payments/public/pay-manual-page'

export const Route = createFileRoute('/bayar')({
  component: PayManualPage,
})
