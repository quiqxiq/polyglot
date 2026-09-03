import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DeleteSubscriptionRequest, type Subscription } from '@/gen/v1/subscription_pb'
import { useDeleteSubscriptionMutation } from '@/features/billing/api/use-billing'

interface SubscriptionsDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription | null
}

export function SubscriptionsDeleteDialog({
  open,
  onOpenChange,
  currentRow: subscription,
}: SubscriptionsDeleteDialogProps) {
  const deleteMutation = useDeleteSubscriptionMutation()

  const handleConfirm = async () => {
    if (!subscription) return
    try {
      await deleteMutation.mutateAsync(
        new DeleteSubscriptionRequest({ id: subscription.id })
      )
      toast.success('Langganan dihapus')
      onOpenChange(false)
    } catch (err) {
      // Tampilkan pesan backend apa adanya (guard tagihan).
      toast.error(err instanceof Error ? err.message : 'Gagal menghapus langganan')
    }
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title='Hapus langganan?'
      desc={`Langganan ${
        subscription?.remoteUsername || subscription?.id || ''
      } akan dihapus secara permanen. Penghapusan hanya bisa dilakukan bila langganan tidak memiliki tagihan.`}
      confirmText={deleteMutation.isPending ? 'Menghapus...' : 'Hapus'}
      destructive
      isLoading={deleteMutation.isPending}
      handleConfirm={handleConfirm}
    />
  )
}
