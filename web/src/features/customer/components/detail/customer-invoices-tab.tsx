import { CreditCard, Eye, Printer, Receipt } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { Invoice } from '@/gen/v1/billing_pb'
import { invoiceStatusBadge } from '@/features/invoices/data/constants'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { formatCurrency, formatUnixDate } from '@/lib/format'

interface CustomerInvoicesTabProps {
  invoices: Invoice[]
  isLoading: boolean
  onOpenInvoice: (invoice: Invoice) => void
  onPayInvoice: (invoice: Invoice) => void
  onPrintInvoice: (invoice: Invoice) => void
}

export function CustomerInvoicesTab({
  invoices,
  isLoading,
  onOpenInvoice,
  onPayInvoice,
  onPrintInvoice,
}: CustomerInvoicesTabProps) {
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canBilling = canPermission(permissions, 'billing:manage')

  const unpaidInvoices = invoices.filter((inv) => inv.status !== 'PAID')
  const outstanding = unpaidInvoices.reduce(
    (sum, inv) => sum + Math.max(0, inv.total - inv.paidAmount),
    0
  )

  return (
    <div className='space-y-4'>
      <Card className='shadow-xs'>
        <CardContent className='flex flex-wrap items-center justify-between gap-3 p-4'>
          <div>
            <h2 className='text-sm font-semibold'>Riwayat Tagihan</h2>
            <p className='text-xs text-muted-foreground'>
              {invoices.length} faktur diterbitkan untuk pelanggan ini.
            </p>
          </div>

          <div className='flex items-center gap-6'>
            <div className='text-right'>
              <p className='text-[10px] uppercase tracking-wide text-muted-foreground'>
                Tagihan Belum Dibayar
              </p>
              <p
                className={`font-mono text-lg font-bold ${
                  unpaidInvoices.length > 0
                    ? 'text-amber-600 dark:text-amber-400'
                    : 'text-emerald-600 dark:text-emerald-400'
                }`}
              >
                {formatCurrency(outstanding)}
              </p>
            </div>
            <Badge variant='outline' className='text-[10px]'>
              {unpaidInvoices.length} belum lunas
            </Badge>
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className='space-y-2'>
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className='h-12 w-full' />
          ))}
        </div>
      ) : invoices.length === 0 ? (
        <div className='rounded-xl border border-dashed py-14 text-center'>
          <Receipt className='mx-auto h-10 w-10 text-muted-foreground/40' />
          <p className='mt-3 text-sm font-semibold'>Belum Ada Tagihan</p>
          <p className='mx-auto mt-1 max-w-xs text-xs text-muted-foreground'>
            Faktur bulanan dibuat otomatis oleh sistem billing sesuai siklus
            langganan.
          </p>
        </div>
      ) : (
        <>
          {/* Desktop */}
          <div className='hidden overflow-hidden rounded-md border bg-card md:block'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>No. Faktur</TableHead>
                  <TableHead>Periode</TableHead>
                  <TableHead>Jatuh Tempo</TableHead>
                  <TableHead className='text-right'>Total</TableHead>
                  <TableHead className='text-right'>Sisa</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className='text-right'>Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {invoices.map((invoice) => {
                  const isPaid = invoice.status === 'PAID'
                  const outstandingAmount = Math.max(
                    0,
                    invoice.total - invoice.paidAmount
                  )
                  const badge = invoiceStatusBadge(invoice.status)
                  return (
                    <TableRow key={invoice.id}>
                      <TableCell className='font-mono text-xs font-semibold'>
                        {invoice.invoiceNumber || invoice.id}
                      </TableCell>
                      <TableCell className='text-xs'>{invoice.period || '-'}</TableCell>
                      <TableCell className='text-xs'>
                        {formatUnixDate(invoice.dueDateUnix)}
                      </TableCell>
                      <TableCell className='text-right font-mono text-xs font-semibold'>
                        {formatCurrency(invoice.total)}
                      </TableCell>
                      <TableCell className='text-right font-mono text-xs'>
                        {isPaid ? '-' : formatCurrency(outstandingAmount)}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant='outline'
                          className={`text-[10px] ${badge.className}`}
                        >
                          {badge.label}
                        </Badge>
                      </TableCell>
                      <TableCell className='text-right'>
                        <div className='flex items-center justify-end gap-1'>
                          <Button
                            variant='ghost'
                            size='icon'
                            className='h-7 w-7'
                            title='Detail faktur'
                            onClick={() => onOpenInvoice(invoice)}
                          >
                            <Eye className='h-3.5 w-3.5' />
                          </Button>
                          <Button
                            variant='ghost'
                            size='icon'
                            className='h-7 w-7'
                            title='Cetak faktur'
                            onClick={() => onPrintInvoice(invoice)}
                          >
                            <Printer className='h-3.5 w-3.5' />
                          </Button>
                          {!isPaid && canBilling && (
                            <Button
                              size='sm'
                              className='h-7 gap-1 bg-emerald-600 text-xs text-white hover:bg-emerald-700'
                              onClick={() => onPayInvoice(invoice)}
                            >
                              <CreditCard className='h-3.5 w-3.5' />
                              Bayar
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>

          {/* Mobile */}
          <div className='space-y-3 md:hidden'>
            {invoices.map((invoice) => {
              const isPaid = invoice.status === 'PAID'
              const outstandingAmount = Math.max(
                0,
                invoice.total - invoice.paidAmount
              )
              const badge = invoiceStatusBadge(invoice.status)
              return (
                <Card key={invoice.id} className='shadow-xs'>
                  <CardContent className='space-y-3 p-4'>
                    <div className='flex items-start justify-between gap-3'>
                      <div className='min-w-0'>
                        <p className='truncate font-mono text-sm font-bold'>
                          {invoice.invoiceNumber || invoice.id}
                        </p>
                        <p className='text-xs text-muted-foreground'>
                          {invoice.period || '-'} · Jatuh tempo{' '}
                          {formatUnixDate(invoice.dueDateUnix)}
                        </p>
                      </div>
                      <Badge
                        variant='outline'
                        className={`shrink-0 text-[10px] ${badge.className}`}
                      >
                        {badge.label}
                      </Badge>
                    </div>

                    <div className='flex items-end justify-between gap-3 border-t pt-3'>
                      <div>
                        <p className='text-[10px] uppercase tracking-wide text-muted-foreground'>
                          Total
                        </p>
                        <p className='font-mono text-sm font-bold'>
                          {formatCurrency(invoice.total)}
                        </p>
                        {!isPaid && (
                          <p className='text-[11px] text-muted-foreground'>
                            Sisa {formatCurrency(outstandingAmount)}
                          </p>
                        )}
                      </div>
                      <div className='flex items-center gap-1.5'>
                        <Button
                          variant='outline'
                          size='sm'
                          className='h-7 gap-1 text-xs'
                          onClick={() => onOpenInvoice(invoice)}
                        >
                          <Eye className='h-3.5 w-3.5' />
                          Detail
                        </Button>
                        {!isPaid && canBilling && (
                          <Button
                            size='sm'
                            className='h-7 gap-1 bg-emerald-600 text-xs text-white hover:bg-emerald-700'
                            onClick={() => onPayInvoice(invoice)}
                          >
                            <CreditCard className='h-3.5 w-3.5' />
                            Bayar
                          </Button>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}
