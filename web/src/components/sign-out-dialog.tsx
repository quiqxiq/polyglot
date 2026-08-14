import { useNavigate, useLocation } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useLogoutMutation } from '@/features/auth/api/use-auth'

interface SignOutDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SignOutDialog({ open, onOpenChange }: SignOutDialogProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const logoutMutation = useLogoutMutation()
  const reset = useAuthStore((s) => s.auth.reset)

  const handleSignOut = async () => {
    // Revoke refresh token di server (hapus dari Redis + clear cookie),
    // lalu reset state lokal. Best-effort: kalau server down, logout lokal
    // tetap jalan — access token memory-only hilang dengan reload/redirect.
    try {
      await logoutMutation.mutateAsync()
    } catch {
      // Server down / network error — logout lokal tetap jalan.
    } finally {
      reset()
    }
    // Preserve current location for redirect after sign-in
    const currentPath = location.href
    navigate({
      to: '/sign-in',
      search: { redirect: currentPath },
      replace: true,
    })
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title='Sign out'
      desc='Are you sure you want to sign out? You will need to sign in again to access your account.'
      confirmText='Sign out'
      destructive
      handleConfirm={handleSignOut}
      className='sm:max-w-sm'
    />
  )
}
