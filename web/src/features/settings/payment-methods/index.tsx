import { useState } from 'react'
import { CreditCard, Pencil, Plus, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { PaymentMethod } from '@/gen/v1/billing_pb'
import {
  useDeletePaymentMethodMutation,
  usePaymentMethodsQuery,
  useSavePaymentMethodMutation,
} from '@/features/billing/api/use-billing-ops'

interface FormState {
  id: string
  name: string
  type: string
  isActive: boolean
  bankName: string
  accountNumber: string
  accountHolder: string
  qrisImageUrl: string
  instructions: string
  sortOrder: string
}

const EMPTY_FORM: FormState = {
  id: '',
  name: '',
  type: 'BANK',
  isActive: true,
  bankName: '',
  accountNumber: '',
  accountHolder: '',
  qrisImageUrl: '',
  instructions: '',
  sortOrder: '0',
}

const TYPE_OPTIONS = [
  { value: 'CASH', label: 'Tunai' },
  { value: 'BANK', label: 'Transfer Bank' },
  { value: 'QRIS', label: 'QRIS' },
  { value: 'GATEWAY', label: 'Payment Gateway' },
]

export function SettingsPaymentMethods() {
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<FormState>(EMPTY_FORM)
  const { data: methods = [], isLoading } = usePaymentMethodsQuery(false)
  const save = useSavePaymentMethodMutation()
  const remove = useDeletePaymentMethodMutation()

  function openCreate() {
    setForm(EMPTY_FORM)
    setOpen(true)
  }

  function openEdit(id: string) {
    const existing = methods.find((m) => m.id === id)
    setForm(
      existing
        ? {
            id: existing.id,
            name: existing.name,
            type: existing.type,
            isActive: existing.isActive,
            bankName: existing.bankName,
            accountNumber: existing.accountNumber,
            accountHolder: existing.accountHolder,
            qrisImageUrl: existing.qrisImageUrl,
            instructions: existing.instructions,
            sortOrder: String(existing.sortOrder),
          }
        : { ...EMPTY_FORM, id }
    )
    setOpen(true)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name.trim()) return
    const method = new PaymentMethod({
      id: form.id,
      name: form.name.trim(),
      type: form.type,
      isActive: form.isActive,
      bankName: form.bankName,
      accountNumber: form.accountNumber,
      accountHolder: form.accountHolder,
      qrisImageUrl: form.qrisImageUrl,
      instructions: form.instructions,
      sortOrder: Number(form.sortOrder) || 0,
    })
    await save.mutateAsync(method)
    setOpen(false)
  }

  return (
    <Card>
      <CardHeader className='flex flex-row flex-wrap items-center justify-between gap-3 space-y-0'>
        <div>
          <CardTitle className='flex items-center gap-2 text-base'>
            <CreditCard className='h-4 w-4 text-primary' /> Master Metode Pembayaran
          </CardTitle>
          <CardDescription>
            Kelola kanal pembayaran manual: tunai, transfer bank (dengan nomor
            rekening), QRIS, dan gateway. Metode nonaktif tidak muncul di
            pilihan pembayaran.
          </CardDescription>
        </div>
        <Button size='sm' className='gap-1.5' onClick={openCreate}>
          <Plus className='h-4 w-4' /> Tambah Metode
        </Button>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Nama</TableHead>
              <TableHead>Tipe</TableHead>
              <TableHead>Rekening / QRIS</TableHead>
              <TableHead>Urutan</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className='text-right'>Aksi</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className='py-8 text-center text-sm text-muted-foreground'>
                  Memuat metode pembayaran…
                </TableCell>
              </TableRow>
            ) : methods.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className='py-8 text-center text-sm text-muted-foreground'>
                  Belum ada metode pembayaran.
                </TableCell>
              </TableRow>
            ) : (
              methods.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className='font-medium'>{m.name}</TableCell>
                  <TableCell>
                    <Badge variant='outline'>
                      {TYPE_OPTIONS.find((t) => t.value === m.type)?.label || m.type}
                    </Badge>
                  </TableCell>
                  <TableCell className='text-sm text-muted-foreground'>
                    {m.accountNumber
                      ? `${m.bankName || ''} ${m.accountNumber} ${m.accountHolder ? `a.n. ${m.accountHolder}` : ''}`
                      : m.qrisImageUrl
                        ? 'QRIS tersedia'
                        : '-'}
                  </TableCell>
                  <TableCell>{m.sortOrder}</TableCell>
                  <TableCell>
                    <Badge variant={m.isActive ? 'default' : 'secondary'}>
                      {m.isActive ? 'Aktif' : 'Nonaktif'}
                    </Badge>
                  </TableCell>
                  <TableCell className='text-right'>
                    <div className='flex justify-end gap-1'>
                      <Button variant='ghost' size='icon' onClick={() => openEdit(m.id)}>
                        <Pencil className='h-4 w-4' />
                      </Button>
                      <Button
                        variant='ghost'
                        size='icon'
                        disabled={!m.isActive || remove.isPending}
                        onClick={() => remove.mutate(m.id)}
                      >
                        <Trash2 className='h-4 w-4 text-destructive' />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='sm:max-w-lg'>
          <DialogHeader>
            <DialogTitle>
              {form.id ? 'Ubah Metode Pembayaran' : 'Tambah Metode Pembayaran'}
            </DialogTitle>
            <DialogDescription>
              Data rekening ditampilkan ke pelanggan pada halaman pembayaran.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit} className='grid gap-4 py-2 sm:grid-cols-2'>
            <div className='space-y-1.5 sm:col-span-2'>
              <Label>Nama Metode</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder='Contoh: Transfer BCA'
              />
            </div>
            <div className='space-y-1.5'>
              <Label>Tipe</Label>
              <Select
                value={form.type}
                onValueChange={(value) => setForm({ ...form, type: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TYPE_OPTIONS.map((t) => (
                    <SelectItem key={t.value} value={t.value}>
                      {t.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-1.5'>
              <Label>Urutan Tampil</Label>
              <Input
                type='number'
                value={form.sortOrder}
                onChange={(e) => setForm({ ...form, sortOrder: e.target.value })}
              />
            </div>
            <div className='space-y-1.5'>
              <Label>Nama Bank</Label>
              <Input
                value={form.bankName}
                onChange={(e) => setForm({ ...form, bankName: e.target.value })}
              />
            </div>
            <div className='space-y-1.5'>
              <Label>Nomor Rekening</Label>
              <Input
                value={form.accountNumber}
                onChange={(e) => setForm({ ...form, accountNumber: e.target.value })}
              />
            </div>
            <div className='space-y-1.5 sm:col-span-2'>
              <Label>Atas Nama</Label>
              <Input
                value={form.accountHolder}
                onChange={(e) => setForm({ ...form, accountHolder: e.target.value })}
              />
            </div>
            <div className='space-y-1.5 sm:col-span-2'>
              <Label>URL Gambar QRIS</Label>
              <Input
                value={form.qrisImageUrl}
                onChange={(e) => setForm({ ...form, qrisImageUrl: e.target.value })}
                placeholder='https://…'
              />
            </div>
            <div className='space-y-1.5 sm:col-span-2'>
              <Label>Instruksi Pembayaran</Label>
              <Textarea
                rows={3}
                value={form.instructions}
                onChange={(e) => setForm({ ...form, instructions: e.target.value })}
                placeholder='Contoh: transfer sesuai nominal, lalu unggah bukti.'
              />
            </div>
            <div className='flex items-center gap-2 sm:col-span-2'>
              <Switch
                checked={form.isActive}
                onCheckedChange={(checked) => setForm({ ...form, isActive: checked })}
              />
              <Label>Aktif</Label>
            </div>
            <DialogFooter className='sm:col-span-2'>
              <Button type='button' variant='outline' onClick={() => setOpen(false)}>
                Batal
              </Button>
              <Button type='submit' disabled={save.isPending || !form.name.trim()}>
                {save.isPending ? 'Menyimpan…' : 'Simpan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
