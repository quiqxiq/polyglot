import { useNavigate, useLocation } from '@tanstack/react-router'
import { useI18n } from '@/context/i18n-provider'
import { useAuthStore } from '@/stores/auth-store'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useLogoutMutation } from '@/features/auth/api/use-auth'

interface SignOutDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SignOutDialog({ open, onOpenChange }: SignOutDialogProps) {
  const { t } = useI18n()
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
      title={t('auth.sign_out_confirm_title')}
      desc={t('auth.sign_out_confirm_desc')}
      confirmText={t('auth.sign_out')}
      cancelBtnText={t('common.cancel')}
      destructive
      handleConfirm={handleSignOut}
      className='sm:max-w-sm'
    />
  )
}
