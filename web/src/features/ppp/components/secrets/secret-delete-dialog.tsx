import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeviceStore } from '@/stores/device-store'
import { useDeletePPPSecretMutation } from '../../api/use-ppp-secrets'
import { usePPP } from '../../context/ppp-context'

export function SecretDeleteDialog() {
  const { open, setOpen, currentSecret, setCurrentSecret } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const deleteMutation = useDeletePPPSecretMutation()

  const isOpen = open === 'secret-delete'

  const handleClose = () => {
    setOpen(null)
    setCurrentSecret(null)
  }

  const handleDelete = async () => {
    if (!selectedDeviceId || !currentSecret) return
    await deleteMutation.mutateAsync({
      deviceId: selectedDeviceId,
      rosId: currentSecret.id,
    })
    handleClose()
  }

  return (
    <ConfirmDialog
      open={isOpen}
      onOpenChange={(val) => !val && handleClose()}
      handleConfirm={handleDelete}
      disabled={deleteMutation.isPending}
      title="Delete PPPoE Secret"
      desc={
        <span>
          Are you sure you want to delete PPPoE subscriber secret{' '}
          <strong className="font-mono text-foreground">
            {currentSecret?.name}
          </strong>
          ? This action cannot be undone and will permanently remove this secret from the router.
        </span>
      }
      confirmText="Delete Secret"
      destructive
    />
  )
}
