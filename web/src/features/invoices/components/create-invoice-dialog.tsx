import { useMemo, useState } from 'react'
import { FilePlus2, Plus, Trash2 } from 'lucide-react'
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
import { CreateInvoiceItem } from '@/gen/v1/billing_pb'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { useCreateInvoiceMutation } from '@/features/billing/api/use-billing-ops'

interface ItemState {
  description: string
  quantity: string
  unitPrice: string
}

const EMPTY_ITEM: ItemState = { description: '', quantity: '1', unitPrice: '' }

interface CreateInvoiceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateInvoiceDialog({ open, onOpenChange }: CreateInvoiceDialogProps) {
  const { data: customers = [] } = useCustomersQuery()
  const create = useCreateInvoiceMutation()
  const currentMonth = useMemo(() => {
    const now = new Date()
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  }, [])
  const [customerId, setCustomerId] = useState('')
  const [period, setPeriod] = useState(currentMonth)
  const [dueDate, setDueDate] = useState('')
  const [taxPercent, setTaxPercent] = useState('')
  const [notes, setNotes] = useState('')
  const [items, setItems] = useState<ItemState[]>([{ ...EMPTY_ITEM }])

  const total = items.reduce(
    (sum, it) => sum + Number(it.quantity || 0) * Number(it.unitPrice || 0),
    0
  )

  function updateItem(index: number, patch: Partial<ItemState>) {
    setItems((prev) => prev.map((it, i) => (i === index ? { ...it, ...patch } : it)))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const pbItems = items
      .filter((it) => it.description.trim() && Number(it.unitPrice) > 0)
      .map(
        (it) =>
          new CreateInvoiceItem({
            description: it.description.trim(),
            quantity: Number(it.quantity) || 1,
            unitPrice: Number(it.unitPrice) || 0,
            itemType: 'AD_HOC',
          })
      )
    if (!customerId || pbItems.length === 0) return
    let dueUnix = 0n
    if (dueDate) {
      dueUnix = BigInt(Math.floor(new Date(`${dueDate}T23:59:59`).getTime() / 1000))
    }
    await create.mutateAsync({
      customerId,
      period,
      dueDateUnix: dueUnix,
      taxPercent: Number(taxPercent) || 0,
      items: pbItems,
      notes,
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <FilePlus2 className='h-5 w-5 text-primary' />
            <DialogTitle>Buat Tagihan Manual</DialogTitle>
          </div>
          <DialogDescription>
            Faktur di luar siklus otomatis: biaya instalasi, add-on, atau
            penyesuaian. Tidak ada denda/biaya tambahan otomatis.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className='space-y-4 py-2'>
          <div className='space-y-1.5'>
            <Label>Pelanggan</Label>
            <Select value={customerId} onValueChange={setCustomerId}>
              <SelectTrigger>
                <SelectValue placeholder='Pilih pelanggan' />
              </SelectTrigger>
              <SelectContent>
                {customers.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='grid gap-3 sm:grid-cols-3'>
            <div className='space-y-1.5'>
              <Label>Periode</Label>
              <Input value={period} onChange={(e) => setPeriod(e.target.value)} placeholder='YYYY-MM' />
            </div>
            <div className='space-y-1.5'>
              <Label>Jatuh Tempo</Label>
              <Input type='date' value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
            </div>
            <div className='space-y-1.5'>
              <Label>PPN (%)</Label>
              <Input
                type='number'
                min={0}
                value={taxPercent}
                onChange={(e) => setTaxPercent(e.target.value)}
                placeholder='Default'
              />
            </div>
          </div>

          <div className='space-y-2'>
            <div className='flex items-center justify-between'>
              <Label>Rincian Biaya</Label>
              <Button
                type='button'
                variant='outline'
                size='sm'
                className='gap-1'
                onClick={() => setItems((prev) => [...prev, { ...EMPTY_ITEM }])}
              >
                <Plus className='h-3.5 w-3.5' /> Baris
              </Button>
            </div>
            {items.map((it, index) => (
              <div key={index} className='grid grid-cols-12 gap-2'>
                <Input
                  className='col-span-6'
                  placeholder='Deskripsi'
                  value={it.description}
                  onChange={(e) => updateItem(index, { description: e.target.value })}
                />
                <Input
                  className='col-span-2'
                  type='number'
                  min={1}
                  placeholder='Qty'
                  value={it.quantity}
                  onChange={(e) => updateItem(index, { quantity: e.target.value })}
                />
                <Input
                  className='col-span-3'
                  type='number'
                  min={0}
                  placeholder='Harga'
                  value={it.unitPrice}
                  onChange={(e) => updateItem(index, { unitPrice: e.target.value })}
                />
                <Button
                  type='button'
                  variant='ghost'
                  size='icon'
                  className='col-span-1'
                  disabled={items.length === 1}
                  onClick={() => setItems((prev) => prev.filter((_, i) => i !== index))}
                >
                  <Trash2 className='h-4 w-4 text-destructive' />
                </Button>
              </div>
            ))}
            <p className='text-right text-sm text-muted-foreground'>
              Subtotal: <span className='font-medium text-foreground'>
                Rp{total.toLocaleString('id-ID')}
              </span>
            </p>
          </div>

          <div className='space-y-1.5'>
            <Label>Catatan</Label>
            <Textarea rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} />
          </div>

          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>
              Batal
            </Button>
            <Button type='submit' disabled={create.isPending || !customerId}>
              {create.isPending ? 'Menyimpan…' : 'Terbitkan Tagihan'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
