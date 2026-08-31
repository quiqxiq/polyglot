import { Copy, CreditCard, Eye, MoreHorizontal, Printer } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useInvoices } from './invoices-provider'

interface InvoicesRowActionsProps {
  invoice: Invoice
}

export function InvoicesRowActions({ invoice }: InvoicesRowActionsProps) {
  const { setOpen, setCurrentInvoice } = useInvoices()
  const isPaid = invoice.status === 'PAID'

  const handleCopyCode = () => {
    const code = invoice.manualPaymentCode || invoice.invoiceNumber
    if (code) {
      navigator.clipboard.writeText(code)
      toast.success(`Kode bayar ${code} disalin ke clipboard`)
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' size='icon' className='h-8 w-8 p-0'>
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Buka menu aksi</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-48'>
        {/* Detail Faktur */}
        <DropdownMenuItem
          onClick={() => {
            setCurrentInvoice(invoice)
            setOpen('detail')
          }}
          className='gap-2 cursor-pointer'
        >
          <Eye className='h-4 w-4 text-blue-600' />
          Rincian Faktur
        </DropdownMenuItem>

        {/* Kasir Bayar Langsung */}
        {!isPaid && (
          <DropdownMenuItem
            onClick={() => {
              setCurrentInvoice(invoice)
              setOpen('cashier')
            }}
            className='gap-2 cursor-pointer font-medium text-emerald-600 dark:text-emerald-400'
          >
            <CreditCard className='h-4 w-4' />
            Bayar di Kasir
          </DropdownMenuItem>
        )}

        {/* Cetak */}
        <DropdownMenuItem
          onClick={() => {
            setCurrentInvoice(invoice)
            setOpen('print')
          }}
          className='gap-2 cursor-pointer'
        >
          <Printer className='h-4 w-4 text-slate-600' />
          {isPaid ? 'Cetak Kwitansi' : 'Cetak Faktur'}
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        {/* Salin Kode Bayar */}
        <DropdownMenuItem onClick={handleCopyCode} className='gap-2 cursor-pointer'>
          <Copy className='h-4 w-4 text-muted-foreground' />
          Salin Kode Bayar
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
