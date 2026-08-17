import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useResetHotspotUserCountersMutation } from '../../api/use-hotspot-users'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function UserResetDialog() {
  const { open, setOpen, currentUser, setCurrentUser } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const resetMutation = useResetHotspotUserCountersMutation()

  const isOpen = open === 'user-reset'

  const handleConfirm = async () => {
    if (!currentUser) return
    try {
      await resetMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentUser.id,
      })
      toast.success(`Counters for "${currentUser.name}" have been reset to 0!`)
      setOpen(null)
      setCurrentUser(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to reset user counters')
    }
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => {
        if (!val) {
          setOpen(null)
          setCurrentUser(null)
        }
      }}
      title={`Reset Counters for: ${currentUser?.name}?`}
      desc={
        <>
          This will reset the <strong>uptime</strong> and <strong>bytes transfer</strong> counters of user{' '}
          <code>{currentUser?.name}</code> to zero (<code className='font-mono'>0s / 0 B</code>).
        </>
      }
      confirmText={resetMutation.isPending ? 'Resetting...' : 'Reset Counters'}
      handleConfirm={handleConfirm}
    />
  )
}
