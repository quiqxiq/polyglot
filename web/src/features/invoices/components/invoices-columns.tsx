import { type ColumnDef } from '@tanstack/react-table'
import { CreditCard, Eye } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { type Invoice } from '@/gen/v1/billing_pb'
import { type Customer } from '@/gen/v1/customer_pb'
import { invoiceStatusBadge } from '../data/constants'
import { InvoicesRowActions } from './invoices-row-actions'
import { useInvoices } from './invoices-provider'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

export function useInvoiceColumns(customerMap: Map<string, Customer>): ColumnDef<Invoice>[] {
  const { setOpen, setCurrentInvoice } = useInvoices()

  return [
    {
      accessorKey: 'invoiceNumber',
      header: 'No. Faktur',
      cell: ({ row }) => {
        const inv = row.original
        return (
          <div className='flex flex-col'>
            <span className='font-mono text-xs font-bold text-foreground'>
              {inv.invoiceNumber || inv.id}
            </span>
            {inv.manualPaymentCode && (
              <span className='font-mono text-[10px] text-muted-foreground'>
                Kode: <span className='font-semibold text-primary'>{inv.manualPaymentCode}</span>
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'customerId',
      header: 'Pelanggan',
      cell: ({ row }) => {
        const cust = customerMap.get(row.original.customerId)
        if (!cust) return <span className='text-xs text-muted-foreground'>-</span>
        return (
          <div className='flex flex-col'>
            <span className='text-xs font-semibold text-foreground'>{cust.name}</span>
            <div className='flex items-center gap-1.5 text-[10px] text-muted-foreground'>
              <span className='font-mono'>{cust.customerCode}</span>
              {cust.phone && <span>· {cust.phone}</span>}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: 'period',
      header: 'Periode',
      cell: ({ row }) => (
        <span className='font-mono text-xs font-medium text-foreground'>
          {row.original.period || '-'}
        </span>
      ),
    },
    {
      accessorKey: 'dueDateUnix',
      header: 'Jatuh Tempo',
      cell: ({ row }) => {
        const inv = row.original
        const isOverdue = inv.status === 'OVERDUE'
        return (
          <div className='flex flex-col'>
            <span className={`text-xs font-medium ${isOverdue ? 'text-rose-600 dark:text-rose-400 font-semibold' : 'text-foreground'}`}>
              {formatUnixDate(inv.dueDateUnix)}
            </span>
            {inv.paidAtUnix ? (
              <span className='text-[10px] text-emerald-600 dark:text-emerald-400'>
                Dibayar {formatUnixDate(inv.paidAtUnix)}
              </span>
            ) : null}
          </div>
        )
      },
    },
    {
      accessorKey: 'total',
      header: () => <div className='text-right'>Total Tagihan (IDR)</div>,
      cell: ({ row }) => {
        const inv = row.original
        const outstanding = Math.max(0, inv.total - inv.paidAmount)
        const isPaid = inv.status === 'PAID'

        return (
          <div className='text-right'>
            <span className='font-mono text-xs font-bold text-foreground'>
              {formatCurrency(inv.total)}
            </span>
            {!isPaid && inv.paidAmount > 0 && (
              <p className='font-mono text-[10px] text-muted-foreground'>
                Sisa: <span className='font-semibold text-rose-600'>{formatCurrency(outstanding)}</span>
              </p>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const meta = invoiceStatusBadge(row.original.status)
        return (
          <Badge variant='outline' className={`text-[11px] font-medium ${meta.className}`}>
            {meta.label}
          </Badge>
        )
      },
    },
    {
      id: 'quickActions',
      header: () => <div className='text-right'>Aksi Cepat</div>,
      cell: ({ row }) => {
        const inv = row.original
        const isPaid = inv.status === 'PAID'

        return (
          <div className='flex items-center justify-end gap-1'>
            {!isPaid ? (
              <Button
                variant='outline'
                size='sm'
                onClick={() => {
                  setCurrentInvoice(inv)
                  setOpen('cashier')
                }}
                className='h-7 px-2 text-[11px] gap-1 border-emerald-500/40 text-emerald-600 hover:bg-emerald-500/10 hover:text-emerald-700'
              >
                <CreditCard className='h-3 w-3' />
                Bayar
              </Button>
            ) : (
              <Button
                variant='ghost'
                size='sm'
                onClick={() => {
                  setCurrentInvoice(inv)
                  setOpen('detail')
                }}
                className='h-7 px-2 text-[11px] gap-1'
              >
                <Eye className='h-3 w-3' />
                Rincian
              </Button>
            )}

            <InvoicesRowActions invoice={inv} />
          </div>
        )
      },
    },
  ]
}
