import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteHotspotUserMutation } from '../../api/use-hotspot-users'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function UserDeleteDialog() {
  const { open, setOpen, currentUser, setCurrentUser } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotUserMutation()

  const isOpen = open === 'user-delete'

  const handleConfirm = async () => {
    if (!currentUser) return
    try {
      await deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentUser.id,
      })
      toast.success(`User "${currentUser.name}" deleted successfully!`)
      setOpen(null)
      setCurrentUser(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete hotspot user')
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
      destructive
      title={`Delete User: ${currentUser?.name}?`}
      desc={
        <>
          Are you sure you want to permanently delete user{' '}
          <strong>{currentUser?.name}</strong> from the MikroTik hotspot database?
        </>
      }
      confirmText={deleteMutation.isPending ? 'Deleting...' : 'Delete User'}
      handleConfirm={handleConfirm}
    />
  )
}
