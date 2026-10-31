import { useEffect, useState } from 'react'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useApplyInvoiceDiscountMutation } from '@/features/billing/api/use-billing'

interface InvoiceDiscountDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  invoice: Invoice | null
}

/**
 * Dialog penerapan diskon / promo code manual pada faktur yang belum lunas
 * (planning v2 §1.7 / Fase 5).
 */
export function InvoiceDiscountDialog({
  open,
  onOpenChange,
  invoice,
}: InvoiceDiscountDialogProps) {
  const [discountType, setDiscountType] = useState('PERCENT')
  const [value, setValue] = useState('')
  const [promoCode, setPromoCode] = useState('')
  const mutation = useApplyInvoiceDiscountMutation()

  useEffect(() => {
    if (invoice) {
      setDiscountType(invoice.discountType || 'PERCENT')
      setValue(invoice.discountValue ? String(invoice.discountValue) : '')
      setPromoCode(invoice.promoCode || '')
    }
  }, [invoice])

  const handleSubmit = async () => {
    if (!invoice) return
    const numeric = Number(value)
    if (discountType !== 'NONE' && (!numeric || numeric <= 0)) {
      toast.error('Nilai diskon wajib lebih dari 0')
      return
    }
    try {
      const res = await mutation.mutateAsync({
        invoiceId: invoice.id,
        discountType,
        discountValue: discountType === 'NONE' ? 0 : numeric,
        promoCode: promoCode.trim(),
      })
      toast.success(
        `Diskon diterapkan. Total baru: Rp${Number(res.invoice?.total || 0).toLocaleString('id-ID')}`
      )
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menerapkan diskon')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Diskon / Promo Code</DialogTitle>
          <DialogDescription>
            Terapkan potongan pada faktur {invoice?.invoiceNumber}. Diskon
            dihitung dari subtotal sebelum pajak.
          </DialogDescription>
        </DialogHeader>
        <div className='space-y-3'>
          <div className='space-y-1.5'>
            <label className='text-xs font-medium'>Jenis Diskon</label>
            <Select value={discountType} onValueChange={setDiscountType}>
              <SelectTrigger className='h-9'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='NONE'>Tanpa Diskon</SelectItem>
                <SelectItem value='PERCENT'>Persentase (%)</SelectItem>
                <SelectItem value='AMOUNT'>Nominal (Rp)</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {discountType !== 'NONE' && (
            <div className='space-y-1.5'>
              <label className='text-xs font-medium'>
                Nilai Diskon {discountType === 'PERCENT' ? '(%)' : '(Rp)'}
              </label>
              <Input
                type='number'
                min={0}
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder={discountType === 'PERCENT' ? '10' : '25000'}
              />
            </div>
          )}
          <div className='space-y-1.5'>
            <label className='text-xs font-medium'>Kode Promo (opsional)</label>
            <Input
              value={promoCode}
              onChange={(e) => setPromoCode(e.target.value)}
              placeholder='PROMO10'
              className='font-mono'
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={mutation.isPending}>
            {mutation.isPending ? 'Menyimpan...' : 'Terapkan'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
