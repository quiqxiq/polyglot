import { AlertCircle, CheckCircle2, Clock, Receipt } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Invoice } from '@/gen/v1/billing_pb'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

interface InvoicesSummaryCardsProps {
  invoices: Invoice[]
}

export function InvoicesSummaryCards({ invoices }: InvoicesSummaryCardsProps) {
  let totalBilled = 0
  let totalPaid = 0
  let totalUnpaid = 0
  let totalOverdue = 0
  let countUnpaid = 0
  let countOverdue = 0

  invoices.forEach((inv) => {
    totalBilled += inv.total
    totalPaid += inv.paidAmount
    const outstanding = Math.max(0, inv.total - inv.paidAmount)

    if (inv.status === 'OVERDUE') {
      totalOverdue += outstanding
      countOverdue++
    } else if (inv.status === 'UNPAID' || inv.status === 'PARTIAL') {
      totalUnpaid += outstanding
      countUnpaid++
    }
  })

  return (
    <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
      {/* ── 1. Total Tagihan ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Total Nilai Faktur</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-primary/15 text-primary'>
            <Receipt className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalBilled)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            {invoices.length} tagihan tercatat di sistem
          </p>
        </CardContent>
      </Card>

      {/* ── 2. Terbayar (Lunas) ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Tagihan Terkumpul</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
            <CheckCircle2 className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalPaid)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            Dana pembayaran telah masuk kas
          </p>
        </CardContent>
      </Card>

      {/* ── 3. Belum Lunas (Unpaid) ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Belum Lunas (Unpaid)</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-amber-500/15 text-amber-600 dark:text-amber-400'>
            <Clock className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalUnpaid)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            {countUnpaid} faktur menunggu pembayaran
          </p>
        </CardContent>
      </Card>

      {/* ── 4. Jatuh Tempo (Overdue) ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Jatuh Tempo (Overdue)</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-rose-500/15 text-rose-600 dark:text-rose-400'>
            <AlertCircle className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className={`text-2xl font-bold font-mono tracking-tight ${totalOverdue > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-foreground'}`}>
            {formatCurrency(totalOverdue)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            {countOverdue} faktur melewati jatuh tempo
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
