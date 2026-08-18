import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeviceStore } from '@/stores/device-store'
import { useDeletePPPProfileMutation } from '../../api/use-ppp-profiles'
import { usePPP } from '../../context/ppp-context'

export function ProfileDeleteDialog() {
  const { open, setOpen, currentProfile, setCurrentProfile } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const deleteMutation = useDeletePPPProfileMutation()

  const isOpen = open === 'profile-delete'

  const handleClose = () => {
    setOpen(null)
    setCurrentProfile(null)
  }

  const handleDelete = async () => {
    if (!selectedDeviceId || !currentProfile) return
    await deleteMutation.mutateAsync({
      deviceId: selectedDeviceId,
      rosId: currentProfile.id,
    })
    handleClose()
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => !val && handleClose()}
      handleConfirm={handleDelete}
      disabled={deleteMutation.isPending}
      title="Delete PPP Profile"
      desc={
        <span>
          Are you sure you want to delete PPP profile{' '}
          <strong className="font-mono text-foreground">
            {currentProfile?.name}
          </strong>
          ? If any subscriber secret is currently using this profile, RouterOS may reject this operation until they are reassigned.
        </span>
      }
      confirmText="Delete Profile"
      destructive
    />
  )
}
