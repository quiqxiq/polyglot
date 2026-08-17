import { DollarSign, Ticket, Calendar } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface ReportsSummaryCardsProps {
  totalIncome: number
  totalCount: number
  filterLabel: string
}

export function ReportsSummaryCards({
  totalIncome,
  totalCount,
  filterLabel,
}: ReportsSummaryCardsProps) {
  return (
    <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Total Income</CardTitle>
          <DollarSign className='size-4 text-emerald-500' />
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400'>
            Rp {totalIncome.toLocaleString('id-ID')}
          </div>
          <p className='text-xs text-muted-foreground mt-1'>
            Total recorded income ({filterLabel})
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Vouchers Sold</CardTitle>
          <Ticket className='size-4 text-primary' />
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold font-mono'>
            {totalCount.toLocaleString('id-ID')}
          </div>
          <p className='text-xs text-muted-foreground mt-1'>
            Expired & recorded vouchers ({filterLabel})
          </p>
        </CardContent>
      </Card>

      <Card className='hidden lg:block'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Active Filter</CardTitle>
          <Calendar className='size-4 text-muted-foreground' />
        </CardHeader>
        <CardContent>
          <div className='text-lg font-semibold capitalize'>
            {filterLabel}
          </div>
          <p className='text-xs text-muted-foreground mt-1'>
            Sales script logs on MikroTik
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
