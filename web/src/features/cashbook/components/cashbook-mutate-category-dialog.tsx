import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { type CashCategory } from '@/gen/v1/cashbook_pb'
import { useSaveCashCategoryMutation } from '../api/use-cashbook'
import { cashCategorySchema, type CashCategoryFormValues } from '../data/schema'

interface CashbookMutateCategoryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentCategory?: CashCategory | null
}

export function CashbookMutateCategoryDialog({
  open,
  onOpenChange,
  currentCategory,
}: CashbookMutateCategoryDialogProps) {
  const saveCategory = useSaveCashCategoryMutation()
  const isEdit = Boolean(currentCategory?.id)

  const form = useForm<CashCategoryFormValues>({
    resolver: zodResolver(cashCategorySchema),
    defaultValues: {
      name: '',
      type: 'INCOME',
      isActive: true,
    },
  })

  useEffect(() => {
    if (currentCategory) {
      form.reset({
        id: currentCategory.id,
        name: currentCategory.name,
        type: (currentCategory.type as 'INCOME' | 'EXPENSE') || 'INCOME',
        isActive: currentCategory.isActive,
      })
    } else {
      form.reset({
        name: '',
        type: 'INCOME',
        isActive: true,
      })
    }
  }, [currentCategory, form])

  const onSubmit = async (values: CashCategoryFormValues) => {
    try {
      await saveCategory.mutateAsync({
        id: currentCategory?.id || '',
        name: values.name,
        type: values.type,
        isActive: values.isActive,
      })
      toast.success(
        isEdit
          ? 'Kategori kas berhasil diperbarui'
          : 'Kategori kas baru berhasil ditambahkan'
      )
      onOpenChange(false)
    } catch (err: any) {
      toast.error('Gagal menyimpan kategori kas', {
        description: err?.message || 'Terjadi kesalahan pada server',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Ubah Kategori Kas' : 'Tambah Kategori Kas'}</DialogTitle>
          <DialogDescription>
            Kategori pos penerimaan pemasukan atau pengeluaran operasional.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4 py-2'>
            {/* Nama Kategori */}
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nama Kategori</FormLabel>
                  <FormControl>
                    <Input placeholder='Contoh: Biaya Listrik & Sewa, Pendapatan Pasang Baru' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Tipe Arus Kas */}
            <FormField
              control={form.control}
              name='type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Jenis Arus Kas</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Pilih jenis arus kas' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value='INCOME'>Pemasukan / Pendapatan (INCOME)</SelectItem>
                      <SelectItem value='EXPENSE'>Pengeluaran / Biaya (EXPENSE)</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Status Aktif */}
            <FormField
              control={form.control}
              name='isActive'
              render={({ field }) => (
                <FormItem className='flex items-center justify-between rounded-lg border p-3'>
                  <div className='space-y-0.5'>
                    <FormLabel>Status Aktif</FormLabel>
                    <p className='text-xs text-muted-foreground'>
                      Kategori aktif dapat dipilih saat pencatatan transaksi
                    </p>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <DialogFooter className='pt-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
                disabled={saveCategory.isPending}
              >
                Batal
              </Button>
              <Button type='submit' disabled={saveCategory.isPending}>
                {saveCategory.isPending ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Buat Kategori'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
