import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { type Customer, DeleteCustomerRequest } from '@/gen/v1/customer_pb'
import { useDeleteCustomerMutation } from '../api/use-customer'

type CustomersDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Customer | null
}

export function CustomersDeleteDialog({
  open,
  onOpenChange,
  currentRow,
}: CustomersDeleteDialogProps) {
  const deleteMutation = useDeleteCustomerMutation()

  const handleDelete = async () => {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(
        new DeleteCustomerRequest({ id: currentRow.id })
      )
      toast.success('Pelanggan dihapus')
      onOpenChange(false)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Gagal menghapus pelanggan'
      toast.error(errorMessage)
    }
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={(v) => onOpenChange(v)}
      handleConfirm={handleDelete}
      title='Hapus pelanggan?'
      desc={
        <>
          Kamu akan menghapus pelanggan <strong>{currentRow?.name}</strong>{' '}
          beserta seluruh langganan, akun router, dan invoice terkait secara permanen. Tindakan ini tidak dapat dibatalkan.
        </>
      }
      confirmText='Delete'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
