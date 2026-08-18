import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeviceStore } from '@/stores/device-store'
import { useKickPPPActiveSessionMutation } from '../../api/use-ppp-active'
import { usePPP } from '../../context/ppp-context'

export function ActiveKickDialog() {
  const { open, setOpen, currentActiveSession, setCurrentActiveSession } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const kickMutation = useKickPPPActiveSessionMutation()

  const isOpen = open === 'active-kick'

  const handleClose = () => {
    setOpen(null)
    setCurrentActiveSession(null)
  }

  const handleKick = async () => {
    if (!selectedDeviceId || !currentActiveSession) return
    await kickMutation.mutateAsync({
      deviceId: selectedDeviceId,
      rosId: currentActiveSession.id,
    })
    handleClose()
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => !val && handleClose()}
      handleConfirm={handleKick}
      disabled={kickMutation.isPending}
      title="Disconnect Active PPPoE Session"
      desc={
        <span>
          Are you sure you want to forcibly disconnect the active connection for subscriber{' '}
          <strong className="font-mono text-foreground">
            {currentActiveSession?.name}
          </strong>{' '}
          (IP: <span className="font-mono">{currentActiveSession?.address || 'N/A'}</span>)? The subscriber's CPE will be logged out immediately.
        </span>
      }
      confirmText="Disconnect Session"
      destructive
    />
  )
}
