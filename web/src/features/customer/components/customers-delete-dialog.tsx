import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DeleteCustomerRequest } from '@/gen/v1/customer_pb'
import { useDeleteCustomerMutation } from '../api/use-customer'
import { useCustomers } from './customers-provider'

export function CustomersDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useCustomers()
  const deleteMutation = useDeleteCustomerMutation()

  const handleDelete = async () => {
    if (!currentRow) return

    try {
      await deleteMutation.mutateAsync(
        new DeleteCustomerRequest({ id: currentRow.id })
      )
      toast.success('Pelanggan dihapus')
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Gagal menghapus pelanggan'
      toast.error(errorMessage)
    }
  }

  return (
    <ConfirmDialog
      open={open === 'delete'}
      onOpenChange={() => setOpen(null)}
      handleConfirm={handleDelete}
      title='Hapus pelanggan?'
      desc={
        <>
          Kamu akan menghapus pelanggan <strong>{currentRow?.name}</strong>{' '}
          secara permanen. Tindakan ini tidak dapat dibatalkan.
        </>
      }
      confirmText='Delete'
      destructive
      isLoading={deleteMutation.isPending}
    />
  )
}
