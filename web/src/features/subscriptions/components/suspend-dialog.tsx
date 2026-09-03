import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  SuspendSubscriptionRequest,
  type Subscription,
} from '@/gen/v1/subscription_pb'
import { useSuspendSubscriptionMutation } from '@/features/billing/api/use-billing'

interface SuspendDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription | null
}

export function SuspendDialog({ open, onOpenChange, currentRow: subscription }: SuspendDialogProps) {
  const suspend = useSuspendSubscriptionMutation()
  const [reason, setReason] = useState('')

  const handleSubmit = async () => {
    if (!subscription || !reason.trim()) return
    try {
      await suspend.mutateAsync(
        new SuspendSubscriptionRequest({
          subscriptionId: subscription.id,
          reason: reason.trim(),
        })
      )
      toast.success('Langganan berhasil ditangguhkan')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menangguhkan langganan')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Tangguhkan Langganan</DialogTitle>
          <DialogDescription>
            Suspend langganan{' '}
            <span className='font-mono text-xs'>
              {subscription?.remoteUsername || subscription?.id}
            </span>
            . User router akan dinonaktifkan sampai di-resume.
          </DialogDescription>
        </DialogHeader>

        <div className='flex flex-col gap-2 py-2'>
          <Label>Alasan (wajib)</Label>
          <Textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder='Contoh: Tunggu pembayaran tagihan bulan ini'
            rows={3}
          />
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='destructive'
            onClick={handleSubmit}
            disabled={!reason.trim() || suspend.isPending}
          >
            {suspend.isPending ? 'Memproses...' : 'Suspend'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
