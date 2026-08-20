import { useMemo } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useHotspotReportsQuery } from '@/features/reports/api/use-reports'

interface SalesChartProps {
  deviceId: string
}

function formatIDR(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatK(amount: number): string {
  if (amount >= 1_000_000) {
    return `${(amount / 1_000_000).toFixed(1)}M`
  }
  if (amount >= 1_000) {
    return `${Math.round(amount / 1_000)}k`
  }
  return String(amount)
}

export function SalesChart({ deviceId }: SalesChartProps) {
  const reportsQuery = useHotspotReportsQuery(deviceId, '', '', '', Boolean(deviceId))
  const reportsData = reportsQuery.data

  const chartData = useMemo(() => {
    if (!reportsData?.reports || reportsData.reports.length === 0) {
      return []
    }

    // Kelompokkan pendapatan per tanggal
    const map = new Map<string, { date: string; income: number; count: number }>()

    for (const r of reportsData.reports) {
      const dateKey = r.date || 'Lainnya'
      const existing = map.get(dateKey) || { date: dateKey, income: 0, count: 0 }
      existing.income += r.price
      existing.count += 1
      map.set(dateKey, existing)
    }

    // Ambil maksimal 14 hari terakhir untuk tampilan grafik yang rapi
    return Array.from(map.values()).slice(-14)
  }, [reportsData])

  return (
    <Card className='col-span-4 shadow-xs'>
      <CardHeader className='flex flex-row items-center justify-between pb-2'>
        <div>
          <CardTitle className='text-base font-semibold'>Tren Penjualan Voucher</CardTitle>
          <CardDescription>
            {reportsData?.total
              ? `Total ${reportsData.total} voucher terjual (${formatIDR(reportsData.totalIncome)})`
              : 'Visualisasi harian omset penjualan voucher Hotspot'}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className='pt-4'>
        {reportsQuery.isLoading ? (
          <Skeleton className='h-[280px] w-full' />
        ) : chartData.length === 0 ? (
          <div className='flex h-[280px] flex-col items-center justify-center rounded-md border border-dashed text-center text-sm text-muted-foreground'>
            <p className='font-medium'>Belum ada rekaman penjualan voucher di router ini.</p>
            <p className='text-xs mt-1 text-muted-foreground/80'>
              Voucher yang terpakai dan terjual di Mikhmon Hotspot akan muncul otomatis di sini.
            </p>
          </div>
        ) : (
          <ResponsiveContainer width='100%' height={280}>
            <BarChart data={chartData} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
              <CartesianGrid strokeDasharray='3 3' vertical={false} className='stroke-muted' />
              <XAxis
                dataKey='date'
                stroke='#888888'
                fontSize={12}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke='#888888'
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(val) => formatK(Number(val))}
              />
              <Tooltip
                content={({ active, payload }) => {
                  if (active && payload && payload.length) {
                    const data = payload[0].payload
                    return (
                      <div className='rounded-lg border bg-background p-2.5 shadow-md text-xs'>
                        <p className='font-medium text-foreground'>{data.date}</p>
                        <p className='mt-1 text-emerald-600 font-semibold'>
                          Omset: {formatIDR(data.income)}
                        </p>
                        <p className='text-muted-foreground'>{data.count} voucher terjual</p>
                      </div>
                    )
                  }
                  return null
                }}
              />
              <Bar
                dataKey='income'
                fill='currentColor'
                radius={[4, 4, 0, 0]}
                className='fill-primary'
              />
            </BarChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
