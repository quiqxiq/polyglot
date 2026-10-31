import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DeletePlanRequest } from '@/gen/v1/plan_pb'
import { useDeviceStore } from '@/stores/device-store'
import { useDeletePlanMutation } from '../api/use-plans'
import { usePlans } from './plans-provider'

export function PlansDeleteDialog() {
  const { t } = useI18n()
  const { open, setOpen, currentRow, setCurrentRow } = usePlans()
  const deleteMutation = useDeletePlanMutation()
  const selectedDeviceId = useDeviceStore((s) => s.selectedDeviceId)

  const handleDelete = async () => {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(
        new DeletePlanRequest({
          id: currentRow.id,
          deviceId: selectedDeviceId || '',
        })
      )
      toast.success(t('alerts.delete_success'))
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : t('alerts.action_failed')
      toast.error(errorMessage)
    }
  }

  return (
    <ConfirmDialog
      open={open === 'delete'}
      onOpenChange={() => setOpen(null)}
      handleConfirm={handleDelete}
      title={t('billing.plans_delete_title')}
      desc={
        <>
          {t('billing.plans_delete_confirm')}
          {currentRow?.name && (
            <div className='mt-2 font-mono text-xs font-semibold'>
              {currentRow.name}
            </div>
          )}
        </>
      }
      confirmText={t('common.delete')}
      cancelBtnText={t('common.cancel')}
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
