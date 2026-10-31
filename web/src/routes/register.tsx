import { createFileRoute } from '@tanstack/react-router'
import { RegisterPublicPage } from '@/features/registration/public/register-public-page'

export const Route = createFileRoute('/register')({
  component: RegisterPublicPage,
})
