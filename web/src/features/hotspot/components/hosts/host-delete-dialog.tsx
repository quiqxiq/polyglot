import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useRemoveHotspotHostMutation } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function HostDeleteDialog() {
  const { open, setOpen, currentHost, setCurrentHost } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const removeMutation = useRemoveHotspotHostMutation()

  const isOpen = open === 'host-delete'

  const handleConfirm = async () => {
    if (!currentHost) return
    try {
      await removeMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentHost.id,
      })
      toast.success(`Host entry "${currentHost.address || currentHost.macAddress}" removed!`)
      setOpen(null)
      setCurrentHost(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to remove host')
    }
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => {
        if (!val) {
          setOpen(null)
          setCurrentHost(null)
        }
      }}
      destructive
      title={`Remove Host Entry: ${currentHost?.macAddress}?`}
      desc={
        <>
          Remove host entry <strong>{currentHost?.macAddress}</strong> ({currentHost?.address}) from the hotspot host table? <br />
          This will clear the host binding from MikroTik memory.
        </>
      }
      confirmText={removeMutation.isPending ? 'Removing...' : 'Remove Host'}
      handleConfirm={handleConfirm}
    />
  )
}
