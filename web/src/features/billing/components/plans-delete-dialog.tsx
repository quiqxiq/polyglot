import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DeletePlanRequest } from '@/gen/v1/billing_pb'
import { useDeletePlanMutation } from '../api/use-plans'
import { usePlans } from './plans-provider'

export function PlansDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = usePlans()
  const deleteMutation = useDeletePlanMutation()

  const handleDelete = async () => {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(
        new DeletePlanRequest({ id: currentRow.id })
      )
      toast.success('Paket dihapus')
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      // Surface the backend message verbatim (e.g. plan masih dipakai
      // langganan aktif) so the admin sees the real blocker.
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
      confirmText='Delete'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
