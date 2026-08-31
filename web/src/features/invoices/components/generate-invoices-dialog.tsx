import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { CalendarPlus, Sparkles } from 'lucide-react'
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
import { GenerateInvoicesRequest } from '@/gen/v1/billing_pb'
import { useGenerateInvoicesMutation } from '../api/use-invoices'
import { generateInvoicesSchema, type GenerateInvoicesFormValues } from '../data/schema'

interface GenerateInvoicesDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function GenerateInvoicesDialog({
  open,
  onOpenChange,
}: GenerateInvoicesDialogProps) {
  const generateInvoices = useGenerateInvoicesMutation()

  // Default periode: bulan berjalan YYYY-MM
  const now = new Date()
  const currentMonthStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`

  const form = useForm<GenerateInvoicesFormValues>({
    resolver: zodResolver(generateInvoicesSchema),
    defaultValues: {
      period: currentMonthStr,
    },
  })

  const onSubmit = async (values: GenerateInvoicesFormValues) => {
    try {
      const res = await generateInvoices.mutateAsync(
        new GenerateInvoicesRequest({
          period: values.period || currentMonthStr,
        })
      )

      toast.success('Generator Tagihan Selesai Diproses!', {
        description: `Berhasil menerbitkan ${res.created} tagihan baru (${res.skipped} pelanggan dilewati karena sudah memiliki tagihan periode ini).`,
      })
      onOpenChange(false)
    } catch (err: any) {
      toast.error('Gagal menjalankan generator tagihan', {
        description: err?.message || 'Terjadi kesalahan pada backend billing engine',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <CalendarPlus className='h-5 w-5 text-primary' />
            <DialogTitle>Generate Tagihan Bulanan</DialogTitle>
          </div>
          <DialogDescription>
            Terbitkan faktur tagihan bulanan secara otomatis untuk seluruh pelanggan dengan langganan aktif.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4 py-2'>
            <FormField
              control={form.control}
              name='period'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Periode Penagihan (Format: YYYY-MM)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Contoh: 2026-09'
                      className='font-mono'
                      {...field}
                    />
                  </FormControl>
                  <p className='text-xs text-muted-foreground'>
                    Kosongkan atau biarkan default untuk men-generate bulan berjalan ({currentMonthStr}).
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground space-y-1.5'>
              <p className='font-semibold text-foreground flex items-center gap-1.5'>
                <Sparkles className='h-3.5 w-3.5 text-primary' /> Informasi Billing Engine:
              </p>
              <ul className='list-disc list-inside space-y-1'>
                <li>Hanya pelanggan dengan langganan berstatus AKTIF yang akan diproses.</li>
                <li>Pelanggan yang sudah memiliki tagihan pada periode terpilih akan otomatis dilewati (anti-duplikasi).</li>
                <li>Nominal tagihan dihitung dari harga paket layanan ditambah pajak dan penyesuaian custom price.</li>
              </ul>
            </div>

            <DialogFooter className='pt-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
                disabled={generateInvoices.isPending}
              >
                Batal
              </Button>
              <Button type='submit' disabled={generateInvoices.isPending} className='gap-1.5'>
                <CalendarPlus className='h-4 w-4' />
                {generateInvoices.isPending ? 'Menjalankan Generator...' : 'Generate Tagihan Sekarang'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
