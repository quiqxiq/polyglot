import { createFileRoute } from '@tanstack/react-router'
import { IsolatePPPoEView } from '@/features/portal/components/isolate-pppoe-view'

export const Route = createFileRoute('/portal/isolate/pppoe')({
  component: IsolatePPPoEView,
})
