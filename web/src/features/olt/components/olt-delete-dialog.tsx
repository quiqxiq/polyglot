import type { OltDevice } from '@/gen/v1/olt_pb'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { useDeleteOltMutation } from '../api/use-olt'

interface OltDeleteDialogProps {
  olt: OltDevice
  onClose: () => void
  onDeleted?: () => void
}

export function OltDeleteDialog({
  olt,
  onClose,
  onDeleted,
}: OltDeleteDialogProps) {
  const { t } = useI18n()
  const deleteMutation = useDeleteOltMutation()

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(olt.id)
      toast.success(t('olt.delete_success'))
      onClose()
      onDeleted?.()
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t('alerts.action_failed')
      )
    }
  }

  return (
    <ConfirmDialog
      open
      onOpenChange={(open) => {
        if (!open && !deleteMutation.isPending) onClose()
      }}
      title={t('olt.delete_title')}
      desc={t('olt.delete_confirm', { name: olt.name || olt.host })}
      confirmText={t('olt.delete')}
      destructive
      isLoading={deleteMutation.isPending}
      handleConfirm={handleDelete}
    />
  )
}
