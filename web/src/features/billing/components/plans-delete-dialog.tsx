import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DeletePlanRequest } from '@/gen/v1/plan_pb'
import { useDeviceStore } from '@/stores/device-store'
import { useDeletePlanMutation } from '../api/use-plans'
import { usePlans } from './plans-provider'

export function PlansDeleteDialog() {
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
      toast.success('Paket berhasil dihapus')
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Gagal menghapus paket'
      toast.error(errorMessage)
    }
  }

  return (
    <ConfirmDialog
      open={open === 'delete'}
      onOpenChange={() => setOpen(null)}
      handleConfirm={handleDelete}
      title='Hapus paket?'
      desc={
        <>
          Kamu akan menghapus paket layanan{' '}
          <strong>{currentRow?.name}</strong> secara permanen. Tindakan ini
          tidak dapat dibatalkan.
        </>
      }
      confirmText='Hapus'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
