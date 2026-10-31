import { useState } from 'react'
import { Wallet } from 'lucide-react'
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
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { usePaymentMethodsQuery } from '@/features/billing/api/use-billing-ops'
import { usePayDebtMutation } from '../api/use-debts'
import type { Debt } from '@/gen/v1/debt_pb'

interface DebtPayDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  debt: Debt | null
  customerName: string
}

function formatIDR(value: number | undefined) {
  return `Rp${Number(value || 0).toLocaleString('id-ID')}`
}

export function DebtPayDialog({ open, onOpenChange, debt, customerName }: DebtPayDialogProps) {
  const [amount, setAmount] = useState(String(debt?.remainingAmount ?? ''))
  const [methodId, setMethodId] = useState('')
  const [notes, setNotes] = useState('')
  const pay = usePayDebtMutation()
  const { data: methods = [] } = usePaymentMethodsQuery(true)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!debt) return
    const value = Number(amount)
    if (!value || value <= 0) return
    await pay.mutateAsync({
      debtId: debt.id,
      amount: value,
      paymentMethodId: methodId || undefined,
      notes,
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <Wallet className='h-5 w-5 text-primary' />
            <DialogTitle>Terima Pembayaran Hutang</DialogTitle>
          </div>
          <DialogDescription>
            {customerName ? `${customerName} — ` : ''}sisa hutang{' '}
            {formatIDR(debt?.remainingAmount)}. Pembayaran otomatis masuk ke
            buku kas.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className='space-y-4 py-2'>
          <div className='space-y-2'>
            <Label>Jumlah Bayar (Rp)</Label>
            <Input
              type='number'
              min={1}
              max={Number(debt?.remainingAmount || 0)}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Boleh cicil; maksimal sebesar sisa hutang. Tidak ada denda
              keterlambatan.
            </p>
          </div>

          <div className='space-y-2'>
            <Label>Metode Pembayaran (opsional)</Label>
            <Select value={methodId} onValueChange={setMethodId}>
              <SelectTrigger>
                <SelectValue placeholder='Pilih metode' />
              </SelectTrigger>
              <SelectContent>
                {methods.map((m) => (
                  <SelectItem key={m.id} value={m.id}>
                    {m.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-2'>
            <Label>Catatan</Label>
            <Textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={2}
              placeholder='Contoh: transfer BCA'
            />
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
              disabled={pay.isPending}
            >
              Batal
            </Button>
            <Button type='submit' disabled={pay.isPending || !amount}>
              {pay.isPending ? 'Menyimpan…' : 'Simpan Pembayaran'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
