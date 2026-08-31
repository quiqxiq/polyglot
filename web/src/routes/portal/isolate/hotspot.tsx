import { createFileRoute } from '@tanstack/react-router'
import { IsolateHotspotView } from '@/features/portal/components/isolate-hotspot-view'

export const Route = createFileRoute('/portal/isolate/hotspot')({
  component: IsolateHotspotView,
})
