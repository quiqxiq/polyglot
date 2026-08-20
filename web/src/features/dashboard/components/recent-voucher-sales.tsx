import { Link } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useHotspotReportsQuery } from '@/features/reports/api/use-reports'
import { ArrowUpRight, Ticket } from 'lucide-react'

interface RecentVoucherSalesProps {
  deviceId: string
}

function formatIDR(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function RecentVoucherSales({ deviceId }: RecentVoucherSalesProps) {
  const reportsQuery = useHotspotReportsQuery(deviceId, '', '', '', Boolean(deviceId))
  const reports = reportsQuery.data?.reports ?? []

  // Ambil 6 data penjualan voucher paling baru
  const recentReports = reports.slice(-6).reverse()

  return (
    <Card className='col-span-3 shadow-xs flex flex-col justify-between'>
      <div>
        <CardHeader className='flex flex-row items-center justify-between pb-3'>
          <div>
            <CardTitle className='text-base font-semibold'>Penjualan Voucher Terkini</CardTitle>
            <CardDescription>Riwayat voucher yang baru saja login/terjual</CardDescription>
          </div>
          <Button asChild size='sm' variant='ghost' className='h-8 gap-1 text-xs'>
            <Link to='/reports'>
              Lihat Semua <ArrowUpRight className='size-3.5' />
            </Link>
          </Button>
        </CardHeader>
        <CardContent className='pt-1'>
          {reportsQuery.isLoading ? (
            <div className='space-y-3'>
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className='flex items-center justify-between gap-2'>
                  <Skeleton className='h-9 w-9 rounded-full' />
                  <div className='flex-1 space-y-1'>
                    <Skeleton className='h-4 w-28' />
                    <Skeleton className='h-3 w-20' />
                  </div>
                  <Skeleton className='h-4 w-16' />
                </div>
              ))}
            </div>
          ) : recentReports.length === 0 ? (
            <div className='flex h-44 flex-col items-center justify-center text-center text-xs text-muted-foreground'>
              <Ticket className='size-8 text-muted-foreground/40 mb-1.5' />
              <p>Belum ada riwayat transaksi voucher.</p>
            </div>
          ) : (
            <div className='space-y-3.5'>
              {recentReports.map((report, idx) => (
                <div key={report.id || idx} className='flex items-center justify-between gap-3 text-xs'>
                  <div className='flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary font-mono text-[11px] font-bold'>
                    {report.username?.charAt(0)?.toUpperCase() || 'V'}
                  </div>
                  <div className='min-w-0 flex-1'>
                    <div className='flex items-center gap-1.5'>
                      <p className='truncate font-medium text-foreground'>{report.username}</p>
                      {report.profile && (
                        <Badge variant='outline' className='px-1 py-0 text-[10px] font-normal truncate max-w-24'>
                          {report.profile}
                        </Badge>
                      )}
                    </div>
                    <p className='text-muted-foreground text-[11px]'>
                      {report.date} {report.time ? `• ${report.time}` : ''}
                    </p>
                  </div>
                  <div className='text-end font-semibold text-emerald-600 dark:text-emerald-400'>
                    +{formatIDR(report.price)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </div>
    </Card>
  )
}
