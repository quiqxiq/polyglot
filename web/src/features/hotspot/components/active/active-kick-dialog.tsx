import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useKickHotspotSessionMutation } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function ActiveKickDialog() {
  const { open, setOpen, currentSession, setCurrentSession } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const kickMutation = useKickHotspotSessionMutation()

  const isOpen = open === 'session-kick'

  const handleConfirm = async () => {
    if (!currentSession) return
    try {
      await kickMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentSession.id,
      })
      toast.success(`Active session for "${currentSession.user}" disconnected!`)
      setOpen(null)
      setCurrentSession(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to kick session')
    }
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => {
        if (!val) {
          setOpen(null)
          setCurrentSession(null)
        }
      }}
      destructive
      title={`Kick Active Session: ${currentSession?.user}?`}
      desc={
        <>
          Are you sure you want to disconnect user <strong>{currentSession?.user}</strong> ({currentSession?.address})? <br />
          The active connection will be terminated immediately.
        </>
      }
      confirmText={kickMutation.isPending ? 'Disconnecting...' : 'Kick User'}
      handleConfirm={handleConfirm}
    />
  )
}
