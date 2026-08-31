import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { ArrowDownLeft, ArrowUpRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AddTransactionRequest } from '@/gen/v1/cashbook_pb'
import {
  useAddCashTransactionMutation,
  useCashAccountsQuery,
  useCashCategoriesQuery,
} from '../api/use-cashbook'
import { addTransactionSchema, type AddTransactionFormValues } from '../data/schema'

interface CashbookTransactionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CashbookTransactionDialog({
  open,
  onOpenChange,
}: CashbookTransactionDialogProps) {
  const { data: accounts = [] } = useCashAccountsQuery(true)
  const { data: categories = [] } = useCashCategoriesQuery(true)
  const addTransaction = useAddCashTransactionMutation()

  const form = useForm<AddTransactionFormValues>({
    resolver: zodResolver(addTransactionSchema),
    defaultValues: {
      accountId: accounts[0]?.id || '',
      categoryId: '',
      direction: 'OUT',
      amount: 0,
      description: '',
    },
  })

  const currentDirection = form.watch('direction')
  const filteredCategories = categories.filter((c) => c.type === (currentDirection === 'IN' ? 'INCOME' : 'EXPENSE'))

  const onSubmit = async (values: AddTransactionFormValues) => {
    try {
      await addTransaction.mutateAsync(
        new AddTransactionRequest({
          accountId: values.accountId,
          categoryId: values.categoryId,
          direction: values.direction,
          amount: Number(values.amount),
          description: values.description,
        })
      )
      toast.success(
        values.direction === 'IN'
          ? 'Pemasukan kas berhasil dicatat'
          : 'Pengeluaran kas berhasil dicatat'
      )
      onOpenChange(false)
      form.reset()
    } catch (err: any) {
      toast.error('Gagal mencatat transaksi kas', {
        description: err?.message || 'Terjadi kesalahan pada server',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Catat Mutasi Kas Manual</DialogTitle>
          <DialogDescription>
            Pencatatan kas masuk atau kas keluar operasional perusahaan.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4 py-2'>
            {/* Arah Transaksi */}
            <FormField
              control={form.control}
              name='direction'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Jenis Mutasi Arus Kas</FormLabel>
                  <div className='grid grid-cols-2 gap-3'>
                    <Button
                      type='button'
                      variant={field.value === 'IN' ? 'default' : 'outline'}
                      className={`h-11 justify-center gap-2 ${
                        field.value === 'IN'
                          ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
                          : 'hover:border-emerald-500'
                      }`}
                      onClick={() => {
                        field.onChange('IN')
                        form.setValue('categoryId', '')
                      }}
                    >
                      <ArrowDownLeft className='h-4 w-4' />
                      Kas Masuk (Income)
                    </Button>
                    <Button
                      type='button'
                      variant={field.value === 'OUT' ? 'default' : 'outline'}
                      className={`h-11 justify-center gap-2 ${
                        field.value === 'OUT'
                          ? 'bg-rose-600 hover:bg-rose-700 text-white'
                          : 'hover:border-rose-500'
                      }`}
                      onClick={() => {
                        field.onChange('OUT')
                        form.setValue('categoryId', '')
                      }}
                    >
                      <ArrowUpRight className='h-4 w-4' />
                      Kas Keluar (Expense)
                    </Button>
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Pilihan Rekening */}
            <FormField
              control={form.control}
              name='accountId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Rekening Kas / Bank</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Pilih rekening kas' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {accounts.map((acc) => (
                        <SelectItem key={acc.id} value={acc.id}>
                          {acc.name} ({acc.accountCode}) — {acc.type === 'BANK' ? 'Bank' : 'Kas Fisik'}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Pilihan Kategori */}
            <FormField
              control={form.control}
              name='categoryId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Kategori Transaksi</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Pilih kategori pos kas' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {filteredCategories.map((cat) => (
                        <SelectItem key={cat.id} value={cat.id}>
                          {cat.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Nominal */}
            <FormField
              control={form.control}
              name='amount'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nominal Transaksi (Rp)</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      step='any'
                      placeholder='Contoh: 150000'
                      {...field}
                      value={field.value || ''}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Deskripsi */}
            <FormField
              control={form.control}
              name='description'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Keterangan / Uraian</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: Pembelian perlengkapan kabel ODP, pembayaran listrik kantor...'
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className='pt-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
                disabled={addTransaction.isPending}
              >
                Batal
              </Button>
              <Button type='submit' disabled={addTransaction.isPending}>
                {addTransaction.isPending ? 'Menyimpan...' : 'Simpan Transaksi'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
