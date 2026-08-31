import { ArrowDownLeft, ArrowUpRight, Landmark, TrendingUp, Wallet } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useCashAccountsQuery, useCashBalancesQuery, useCashTransactionsQuery } from '../api/use-cashbook'
import { useCashbook } from './cashbook-provider'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function CashbookSummaryCards() {
  const { filters } = useCashbook()
  const { data: accounts = [] } = useCashAccountsQuery(false)
  const { data: balances = {} } = useCashBalancesQuery(filters.fromUnix, filters.toUnix)
  const { data: transactions = [] } = useCashTransactionsQuery(filters)

  // Hitung saldo kas fisik vs bank
  let totalCash = 0
  let totalBank = 0
  let totalAll = 0

  accounts.forEach((acc) => {
    const bal = balances[acc.id] || 0
    totalAll += bal
    if (acc.type === 'BANK') {
      totalBank += bal
    } else {
      totalCash += bal
    }
  })

  // Hitung pemasukan & pengeluaran dari transaksi periode terpilih
  let totalIncome = 0
  let totalExpense = 0

  transactions.forEach((tx) => {
    if (tx.direction === 'IN') {
      totalIncome += tx.amount
    } else if (tx.direction === 'OUT') {
      totalExpense += tx.amount
    }
  })

  const netCashflow = totalIncome - totalExpense

  return (
    <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
      {/* ── 1. Total Saldo Bersih ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Total Saldo Bersih</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-primary/15 text-primary'>
            <Wallet className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalAll)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground flex items-center gap-2'>
            <span>Kas: {formatCurrency(totalCash)}</span>
            <span>•</span>
            <span>Bank: {formatCurrency(totalBank)}</span>
          </p>
        </CardContent>
      </Card>

      {/* ── 2. Pemasukan Kas ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Pemasukan Kas</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
            <ArrowDownLeft className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalIncome)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground flex items-center gap-1'>
            <TrendingUp className='size-3 text-emerald-500' />
            <span>Total dana masuk periode ini</span>
          </p>
        </CardContent>
      </Card>

      {/* ── 3. Pengeluaran Kas ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Pengeluaran Kas</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-rose-500/15 text-rose-600 dark:text-rose-400'>
            <ArrowUpRight className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono tracking-tight text-foreground'>
            {formatCurrency(totalExpense)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            Total biaya & mutasi keluar
          </p>
        </CardContent>
      </Card>

      {/* ── 4. Net Cashflow ── */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Arus Kas Bersih (Net)</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-blue-500/15 text-blue-600 dark:text-blue-400'>
            <Landmark className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          <div className={`text-2xl font-bold font-mono tracking-tight ${netCashflow < 0 ? 'text-rose-600 dark:text-rose-400' : 'text-foreground'}`}>
            {formatCurrency(netCashflow)}
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            {netCashflow >= 0 ? 'Surplus kas periode ini' : 'Defisit kas periode ini'}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
