import { createFileRoute, Navigate } from '@tanstack/react-router'

export const Route = createFileRoute('/portal/isolate/')({
  component: () => <Navigate to='/portal/isolate/pppoe' replace />,
})
