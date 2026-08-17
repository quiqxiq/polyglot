import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteHotspotProfileMutation } from '../../api/use-hotspot-profiles'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function ProfileDeleteDialog() {
  const { open, setOpen, currentProfile, setCurrentProfile } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const deleteMutation = useDeleteHotspotProfileMutation()

  const isOpen = open === 'profile-delete'

  const handleConfirm = async () => {
    if (!currentProfile) return
    try {
      await deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentProfile.id,
      })
      toast.success(`Profile "${currentProfile.name}" deleted successfully!`)
      setOpen(null)
      setCurrentProfile(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete profile')
    }
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => {
        if (!val) {
          setOpen(null)
          setCurrentProfile(null)
        }
      }}
      destructive
      title={`Delete Profile: ${currentProfile?.name}?`}
      desc={
        <>
          Are you sure you want to delete profile{' '}
          <strong>{currentProfile?.name}</strong> from MikroTik? <br />
          Users currently assigned to this profile may experience connectivity issues.
        </>
      }
      confirmText={deleteMutation.isPending ? 'Deleting...' : 'Delete Profile'}
      handleConfirm={handleConfirm}
    />
  )
}
