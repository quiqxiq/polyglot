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
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useReversePaymentMutation } from '../api/use-payments'

interface ReversePaymentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * Reversal pembayaran (void) — membuat jurnal koreksi kas keluar dan
 * mengembalikan paid_amount invoice (planning v2 Fase 3).
 */
export function ReversePaymentDialog({
  open,
  onOpenChange,
}: ReversePaymentDialogProps) {
  const [paymentID, setPaymentID] = useState('')
  const [reason, setReason] = useState('')
  const reverseMutation = useReversePaymentMutation()

  const handleSubmit = async () => {
    if (!paymentID.trim()) {
      toast.error('Payment ID wajib diisi')
      return
    }
    if (reason.trim().length < 3) {
      toast.error('Alasan reversal minimal 3 karakter')
      return
    }
    try {
      const res = await reverseMutation.mutateAsync({
        paymentId: paymentID.trim(),
        reason: reason.trim(),
      })
      toast.success(
        `Pembayaran ${res.paymentId} dibatalkan; invoice ${res.invoiceId} disesuaikan.`,
      )
      setPaymentID('')
      setReason('')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal reversal pembayaran')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Reversal Pembayaran (Void)</DialogTitle>
          <DialogDescription>
            Batalkan pembayaran yang salah input. Sistem mencatat jurnal koreksi
            kas keluar dan mengembalikan status invoice.
          </DialogDescription>
        </DialogHeader>
        <div className='space-y-3'>
          <Input
            placeholder='Payment ID (mis. pay-xxxx)'
            value={paymentID}
            onChange={(e) => setPaymentID(e.target.value)}
            className='font-mono'
          />
          <Textarea
            placeholder='Alasan reversal'
            rows={3}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
          />
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Batal
          </Button>
          <Button
            variant='destructive'
            onClick={handleSubmit}
            disabled={reverseMutation.isPending}
          >
            {reverseMutation.isPending ? 'Memproses...' : 'Reversal'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
