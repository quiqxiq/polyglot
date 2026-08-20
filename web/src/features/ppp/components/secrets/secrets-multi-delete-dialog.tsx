import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeviceStore } from '@/stores/device-store'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { useBulkDeletePPPSecretsMutation } from '../../api/use-ppp-secrets'

interface SecretsMultiDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  selectedSecrets: PPPSecret[]
  onSuccess?: () => void
}

export function SecretsMultiDeleteDialog({
  open,
  onOpenChange,
  selectedSecrets,
  onSuccess,
}: SecretsMultiDeleteDialogProps) {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const bulkDeleteMutation = useBulkDeletePPPSecretsMutation()

  const handleDelete = async () => {
    if (!selectedDeviceId || selectedSecrets.length === 0) return
    await bulkDeleteMutation.mutateAsync({
      deviceId: selectedDeviceId,
      rosIds: selectedSecrets.map((s) => s.id),
    })
    onOpenChange(false)
    onSuccess?.()
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={bulkDeleteMutation.isPending}
      title={`Delete ${selectedSecrets.length} PPPoE Secrets`}
      desc={
        <span>
          Are you sure you want to permanently delete{' '}
          <strong>{selectedSecrets.length}</strong> selected PPPoE secret(s)?
          This action will remove them from the router and cannot be undone.
        </span>
      }
      confirmText={`Delete ${selectedSecrets.length} Secrets`}
      destructive
    />
  )
}
