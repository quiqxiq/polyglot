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
import { type CashAccount } from '@/gen/v1/cashbook_pb'
import { useSaveCashAccountMutation } from '../api/use-cashbook'
import { cashAccountSchema, type CashAccountFormValues } from '../data/schema'

interface CashbookMutateAccountDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentAccount?: CashAccount | null
}

export function CashbookMutateAccountDialog({
  open,
  onOpenChange,
  currentAccount,
}: CashbookMutateAccountDialogProps) {
  const saveAccount = useSaveCashAccountMutation()
  const isEdit = Boolean(currentAccount?.id)

  const form = useForm<CashAccountFormValues>({
    resolver: zodResolver(cashAccountSchema),
    defaultValues: {
      accountCode: '',
      name: '',
      type: 'CASH',
      isActive: true,
    },
  })

  useEffect(() => {
    if (currentAccount) {
      form.reset({
        id: currentAccount.id,
        accountCode: currentAccount.accountCode,
        name: currentAccount.name,
        type: (currentAccount.type as 'CASH' | 'BANK') || 'CASH',
        isActive: currentAccount.isActive,
      })
    } else {
      form.reset({
        accountCode: '',
        name: '',
        type: 'CASH',
        isActive: true,
      })
    }
  }, [currentAccount, form])

  const onSubmit = async (values: CashAccountFormValues) => {
    try {
      await saveAccount.mutateAsync({
        id: currentAccount?.id || '',
        accountCode: values.accountCode,
        name: values.name,
        type: values.type,
        isActive: values.isActive,
      })
      toast.success(
        isEdit
          ? 'Rekening kas berhasil diperbarui'
          : 'Rekening kas baru berhasil ditambahkan'
      )
      onOpenChange(false)
    } catch (err: any) {
      toast.error('Gagal menyimpan rekening kas', {
        description: err?.message || 'Terjadi kesalahan pada server',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Ubah Rekening Kas' : 'Tambah Rekening Kas / Bank'}</DialogTitle>
          <DialogDescription>
            Konfigurasi akun rekening penampung transaksi kas tunai atau rekening bank.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4 py-2'>
            {/* Kode Akun */}
            <FormField
              control={form.control}
              name='accountCode'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Kode Akun</FormLabel>
                  <FormControl>
                    <Input placeholder='Contoh: 1001-KAS, 1002-BCA' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Nama Rekening */}
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nama Rekening</FormLabel>
                  <FormControl>
                    <Input placeholder='Contoh: Kas Utama Kantor, Rekening BCA Operasional' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Tipe Rekening */}
            <FormField
              control={form.control}
              name='type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Tipe Rekening</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Pilih tipe rekening' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value='CASH'>Kas Fisik / Uang Tunai / Kasir</SelectItem>
                      <SelectItem value='BANK'>Rekening Bank / Transfer</SelectItem>
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
                      Rekening aktif dapat dipilih saat pencatatan transaksi & kasir
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
                disabled={saveAccount.isPending}
              >
                Batal
              </Button>
              <Button type='submit' disabled={saveAccount.isPending}>
                {saveAccount.isPending ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Buat Rekening'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
