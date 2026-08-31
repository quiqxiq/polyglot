import { CalendarPlus, CreditCard } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useInvoices } from './invoices-provider'

export function InvoicesPrimaryButtons() {
  const { setOpen, setCurrentInvoice } = useInvoices()

  return (
    <div className='flex items-center gap-2'>
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
