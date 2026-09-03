import {
  Receipt,
  Copy,
  CreditCard,
  Calendar,
  CheckCircle2,
  Clock,
  AlertCircle,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import type { Invoice } from '@/gen/v1/billing_pb'

interface CustomerInvoicesTabProps {
  invoices: Invoice[]
  isLoading: boolean
  onPayInvoice: (invoice: Invoice) => void
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return null
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function CustomerInvoicesTab({
  invoices,
  isLoading,
  onPayInvoice,
}: CustomerInvoicesTabProps) {
  const unpaidInvoices = invoices.filter((i) => i.status !== 'PAID')
  const totalUnpaid = unpaidInvoices.reduce(
    (sum, inv) => sum + (inv.total - inv.paidAmount),
    0
  )

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} disalin ke clipboard`)
  }

  return (
    <div className='space-y-4 p-6'>
      {/* ─── Header & Financial Summary ─── */}
      <div className='flex flex-col sm:flex-row sm:items-center justify-between gap-3 rounded-xl bg-muted/30 border p-4'>
        <div>
          <h3 className='text-sm font-semibold text-foreground'>Riwayat Tagihan & Pembayaran</h3>
          <p className='text-xs text-muted-foreground mt-0.5'>
            Total {invoices.length} faktur diterbitkan untuk pelanggan ini.
          </p>
        </div>
        {unpaidInvoices.length > 0 ? (
          <div className='flex items-center gap-3 bg-amber-500/10 border border-amber-500/20 rounded-lg px-3 py-2 text-right'>
            <AlertCircle className='h-4 w-4 text-amber-600 dark:text-amber-400 shrink-0' />
            <div>
              <p className='text-[10px] uppercase font-semibold text-amber-700 dark:text-amber-300'>
                {unpaidInvoices.length} Faktur Tertunggak
              </p>
              <p className='text-sm font-bold font-mono text-amber-800 dark:text-amber-200'>
                {formatCurrency(totalUnpaid)}
              </p>
            </div>
          </div>
        ) : (
          <div className='flex items-center gap-2 text-emerald-600 dark:text-emerald-400 text-xs font-semibold'>
            <CheckCircle2 className='h-4 w-4' />
            <span>Semua Faktur Lunas</span>
          </div>
        )}
      </div>

      {isLoading ? (
        <div className='py-16 text-center text-sm text-muted-foreground animate-pulse'>
          Memuat riwayat tagihan...
        </div>
      ) : invoices.length === 0 ? (
        <div className='rounded-xl border border-dashed py-14 text-center'>
          <Receipt className='mx-auto h-10 w-10 text-muted-foreground/40' />
          <p className='mt-3 text-sm font-semibold text-foreground'>Belum Ada Tagihan</p>
          <p className='mt-1 text-xs text-muted-foreground max-w-xs mx-auto'>
            Faktur bulanan akan otomatis dibuat oleh sistem billing sesuai siklus langganan.
          </p>
        </div>
      ) : (
        <div className='space-y-3'>
          {invoices.map((inv) => {
            const isPaid = inv.status === 'PAID'
            const isOverdue = inv.status === 'OVERDUE'
            const outstanding = Math.max(0, inv.total - inv.paidAmount)

            return (
              <div
                key={inv.id}
                className='overflow-hidden rounded-xl border bg-card shadow-xs transition-all hover:border-primary/40'
              >
                <div className='flex items-start justify-between gap-3 p-4 pb-3'>
                  {/* Info Tagihan */}
                  <div className='min-w-0 space-y-1'>
                    <div className='flex items-center gap-2'>
                      <span className='font-mono text-sm font-bold text-foreground'>
                        {inv.invoiceNumber || inv.id}
                      </span>
                      <Button
                        size='icon'
                        variant='ghost'
                        className='h-5 w-5 text-muted-foreground hover:text-foreground'
                        onClick={() => copyToClipboard(inv.invoiceNumber || inv.id, 'Nomor faktur')}
                        title='Salin nomor invoice'
                      >
                        <Copy className='h-3 w-3' />
                      </Button>
                      <Badge
                        variant='outline'
                        className={`text-[10px] px-2 py-0.5 ${
                          isPaid
                            ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
                            : isOverdue
                            ? 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30'
                            : 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30'
                        }`}
                      >
                        {isPaid ? 'Lunas' : isOverdue ? 'Jatuh Tempo' : 'Belum Bayar'}
                      </Badge>
                    </div>

                    <div className='flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground'>
                      <span>Periode: <strong>{inv.period || '-'}</strong></span>
                      {Number(inv.dueDateUnix) > 0 && (
                        <span className='flex items-center gap-1'>
                          <Calendar className='h-3 w-3' />
                          Jatuh tempo: {formatUnixDate(inv.dueDateUnix)}
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Nominal */}
                  <div className='shrink-0 text-right'>
                    <p className='text-sm font-extrabold font-mono text-foreground'>
                      {formatCurrency(inv.total)}
                    </p>
                    {!isPaid && inv.paidAmount > 0 && (
                      <p className='text-[10px] text-muted-foreground'>
                        Terbayar: {formatCurrency(inv.paidAmount)}
                      </p>
                    )}
                    {isPaid && Number(inv.paidAtUnix) > 0 && (
                      <p className='text-[10px] text-emerald-600 dark:text-emerald-400 flex items-center gap-1 justify-end'>
                        <Clock className='h-3 w-3' />
                        Lunas {formatUnixDate(inv.paidAtUnix)}
                      </p>
                    )}
                  </div>
                </div>

                {/* Footer Bar Aksi Pembayaran */}
                {!isPaid && (
                  <div className='flex items-center justify-between gap-3 px-4 py-2.5 border-t bg-muted/20'>
                    <div className='flex items-center gap-2'>
                      <span className='text-xs text-muted-foreground'>Sisa Tagihan:</span>
                      <span className='font-mono font-bold text-xs text-red-600 dark:text-red-400'>
                        {formatCurrency(outstanding)}
                      </span>
                    </div>

                    <div className='flex items-center gap-2'>
                      {inv.manualPaymentCode && (
                        <Button
                          size='sm'
                          variant='ghost'
                          className='h-7 text-xs font-mono gap-1'
                          onClick={() => copyToClipboard(inv.manualPaymentCode, 'Kode bayar')}
                        >
                          <Copy className='h-3 w-3' />
                          {inv.manualPaymentCode}
                        </Button>
                      )}
                      <Button
                        size='sm'
                        className='h-7 gap-1.5 text-xs bg-emerald-600 hover:bg-emerald-700 text-white'
                        onClick={() => onPayInvoice(inv)}
                      >
                        <CreditCard className='h-3.5 w-3.5' />
                        Bayar di Kasir
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
