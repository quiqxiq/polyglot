import { createFileRoute } from '@tanstack/react-router'
import { TrackRegistrationPage } from '@/features/registration/public/track-registration-page'

export const Route = createFileRoute('/register/status')({
  validateSearch: (search: Record<string, unknown>) => ({
    token: typeof search.token === 'string' ? search.token : '',
  }),
  component: TrackRegistrationPage,
})
