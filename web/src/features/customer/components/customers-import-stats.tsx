import { Users, Zap, Radio, Network, FilterX, CheckCircle2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface CustomersImportStatsProps {
  total: number
  pppoe: number
  hotspot: number
  ipBinding: number
  vouchersSkipped: number
  readyCount: number
}

export function CustomersImportStats({
  total,
  pppoe,
  hotspot,
  ipBinding,
  vouchersSkipped,
  readyCount,
}: CustomersImportStatsProps) {
  return (
    <div className='grid grid-cols-2 gap-3 sm:grid-cols-6'>
      {/* 1. Total Terdeteksi */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Total Terdeteksi
          </CardTitle>
          <Users className='h-4 w-4 text-primary' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold'>{total}</div>
        </CardContent>
      </Card>

      {/* 2. PPPoE */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            PPPoE
          </CardTitle>
          <Zap className='h-4 w-4 text-sky-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-sky-600 dark:text-sky-400'>
            {pppoe}
          </div>
        </CardContent>
      </Card>

      {/* 3. Hotspot */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Hotspot
          </CardTitle>
          <Radio className='h-4 w-4 text-amber-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-amber-600 dark:text-amber-400'>
            {hotspot}
          </div>
        </CardContent>
      </Card>

      {/* 4. IP Binding */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            IP Binding
          </CardTitle>
          <Network className='h-4 w-4 text-purple-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-purple-600 dark:text-purple-400'>
            {ipBinding}
          </div>
        </CardContent>
      </Card>

      {/* 5. Voucher Dilewati */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Voucher Lewat
          </CardTitle>
          <FilterX className='h-4 w-4 text-muted-foreground' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-muted-foreground'>
            {vouchersSkipped}
          </div>
        </CardContent>
      </Card>

      {/* 6. Siap Disimpan */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Siap Simpan
          </CardTitle>
          <CheckCircle2 className='h-4 w-4 text-emerald-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-emerald-600 dark:text-emerald-400'>
            {readyCount}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
