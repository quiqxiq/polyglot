import { CalendarPlus, CreditCard, Download, FilePlus2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useExportInvoicesMutation } from '@/features/billing/api/use-billing-ops'
import { useInvoices } from './invoices-provider'

export function InvoicesPrimaryButtons() {
  const { setOpen, setCurrentInvoice, filters } = useInvoices()
  const exportInvoices = useExportInvoicesMutation()

  return (
    <div className='flex flex-wrap items-center gap-2'>
      {/* Tombol Tagihan Manual */}
      <Button
        variant='outline'
        size='sm'
        onClick={() => setOpen('create')}
        className='gap-1.5 h-9'
      >
        <FilePlus2 className='h-4 w-4' />
        Tagihan Manual
      </Button>

      {/* Tombol Generate Tagihan Massal */}
      <Button
        variant='outline'
        size='sm'
        onClick={() => setOpen('generate')}
        className='gap-1.5 h-9'
      >
        <CalendarPlus className='h-4 w-4' />
        Generate Tagihan
      </Button>

      {/* Ekspor daftar tagihan */}
      <Button
        variant='outline'
        size='sm'
        disabled={exportInvoices.isPending}
        onClick={() =>
          exportInvoices.mutate({
            format: 'xlsx',
            filter: {
              customerId: filters.customerId,
              status: filters.status,
              period: filters.period,
              search: filters.search,
              overdueOnly: filters.overdueOnly,
            },
          })
        }
        className='gap-1.5 h-9'
      >
        <Download className='h-4 w-4' />
        Ekspor
      </Button>

      {/* Tombol Kasir POS Cepat */}
      <Button
        size='sm'
        onClick={() => {
          setCurrentInvoice(null)
          setOpen('cashier')
        }}
        className='gap-1.5 h-9'
      >
        <CreditCard className='h-4 w-4' />
        Kasir POS (Bayar Cepat)
      </Button>
    </div>
  )
}
