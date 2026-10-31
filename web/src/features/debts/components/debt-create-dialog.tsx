import { useState } from 'react'
import { HandCoins } from 'lucide-react'
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
import { useCreateDebtMutation } from '../api/use-debts'

interface CustomerOption {
  id: string
  name: string
}

interface DebtCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  customers: CustomerOption[]
}

export function DebtCreateDialog({ open, onOpenChange, customers }: DebtCreateDialogProps) {
  const [customerId, setCustomerId] = useState('')
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [dueDate, setDueDate] = useState('')
  const create = useCreateDebtMutation()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const value = Number(amount)
    if (!customerId || !value || value <= 0) return
    let dueUnix = 0n
    if (dueDate) {
      dueUnix = BigInt(Math.floor(new Date(`${dueDate}T00:00:00`).getTime() / 1000))
    }
    await create.mutateAsync({
      customerId,
      amount: value,
      description,
      dueDateUnix: dueUnix,
      debtDateUnix: BigInt(Math.floor(Date.now() / 1000)),
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <HandCoins className='h-5 w-5 text-primary' />
            <DialogTitle>Catat Hutang Pelanggan</DialogTitle>
          </div>
          <DialogDescription>
            Hutang manual di luar tagihan bulanan. Tidak ada denda atau biaya
            tambahan; pembayaran dicatat ke buku kas sebagai penerimaan.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className='space-y-4 py-2'>
          <div className='space-y-2'>
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

          <div className='space-y-2'>
            <Label>Jumlah Hutang (Rp)</Label>
            <Input
              type='number'
              min={1}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder='Contoh: 150000'
            />
          </div>

          <div className='space-y-2'>
            <Label>Jatuh Tempo (opsional)</Label>
            <Input type='date' value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
          </div>

          <div className='space-y-2'>
            <Label>Keterangan</Label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder='Contoh: pinjaman pemasangan atau sisa tagihan lama'
              rows={3}
            />
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
              disabled={create.isPending}
            >
              Batal
            </Button>
            <Button type='submit' disabled={create.isPending || !customerId || !amount}>
              {create.isPending ? 'Menyimpan…' : 'Simpan Hutang'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
