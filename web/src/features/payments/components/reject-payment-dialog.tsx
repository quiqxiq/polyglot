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
import { Textarea } from '@/components/ui/textarea'
import type { ManualPaymentSubmission } from '@/gen/v1/billing_pb'
import { useRejectPaymentSubmissionMutation } from '../api/use-payments'

interface RejectPaymentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  submission: ManualPaymentSubmission | null
}

export function RejectPaymentDialog({
  open,
  onOpenChange,
  submission,
}: RejectPaymentDialogProps) {
  const [reason, setReason] = useState('')
  const rejectMutation = useRejectPaymentSubmissionMutation()

  const handleReject = async () => {
    if (!submission) return
    if (reason.trim().length < 3) {
      toast.error('Alasan penolakan minimal 3 karakter')
      return
    }
    try {
      await rejectMutation.mutateAsync({ id: submission.id, reason: reason.trim() })
      toast.success('Bukti transfer ditolak')
      setReason('')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menolak bukti transfer')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Tolak Bukti Transfer</DialogTitle>
          <DialogDescription>
            Bukti transfer dari{' '}
            <span className='font-semibold'>{submission?.senderName}</span> akan
            ditolak tanpa mengubah buku kas.
          </DialogDescription>
        </DialogHeader>
        <Textarea
          placeholder='Alasan penolakan (nominal tidak sesuai, bukti tidak jelas, dsb.)'
          rows={3}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
        />
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Batal
          </Button>
          <Button
            variant='destructive'
            onClick={handleReject}
            disabled={rejectMutation.isPending}
          >
            {rejectMutation.isPending ? 'Memproses...' : 'Tolak'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
