import { Badge } from '@/components/ui/badge'
import type { PingMetricPointData } from '@/gen/v1/device_pb'

interface PingAnalyticsTableProps {
  points: PingMetricPointData[]
}

export function PingAnalyticsTable({ points }: PingAnalyticsTableProps) {
  const displayPoints = points.slice(-100).reverse()

  return (
    <div className='border rounded-lg overflow-hidden bg-card shadow-2xs'>
      <div className='max-h-72 overflow-y-auto'>
        <table className='w-full text-xs text-left'>
          <thead className='bg-muted/60 text-muted-foreground sticky top-0 border-b backdrop-blur-xs'>
            <tr>
              <th className='p-2.5 font-medium'>Waktu</th>
              <th className='p-2.5 font-medium'>SEQ</th>
              <th className='p-2.5 font-medium'>Host Target</th>
              <th className='p-2.5 font-medium'>Size</th>
              <th className='p-2.5 font-medium'>TTL</th>
              <th className='p-2.5 font-medium'>Latency</th>
              <th className='p-2.5 font-medium'>Status</th>
            </tr>
          </thead>
          <tbody className='divide-y font-mono text-[11px]'>
            {points.length === 0 ? (
              <tr>
                <td
                  colSpan={7}
                  className='p-6 text-center text-muted-foreground font-sans text-xs'
                >
                  Tidak ada log paket pada rentang ini.
                </td>
              </tr>
            ) : (
              displayPoints.map((p, i) => {
                const isSuccess =
                  p.status === 'connected' ||
                  p.status === 'ok' ||
                  p.packetLoss === 0

                return (
                  <tr key={`${p.timestamp}_${p.seq}_${i}`} className='hover:bg-muted/30 transition-colors'>
                    <td className='p-2.5 text-muted-foreground'>
                      {p.timestamp
                        ? new Date(p.timestamp).toLocaleTimeString()
                        : '-'}
                    </td>
                    <td className='p-2.5 font-semibold text-foreground'>{p.seq}</td>
                    <td className='p-2.5 text-primary'>{p.target}</td>
                    <td className='p-2.5'>{p.size || 56}B</td>
                    <td className='p-2.5'>{p.ttl || '-'}</td>
                    <td className='p-2.5 font-semibold text-foreground'>
                      {p.rttMs > 0 ? `${p.rttMs.toFixed(1)} ms` : 'Timeout'}
                    </td>
                    <td className='p-2.5'>
                      <Badge
                        variant='outline'
                        className={`text-[10px] px-1.5 py-0 capitalize ${
                          isSuccess
                            ? 'bg-emerald-500/10 text-emerald-600 border-emerald-300 dark:border-emerald-800'
                            : 'bg-destructive/10 text-destructive border-destructive/30'
                        }`}
                      >
                        {p.status || (isSuccess ? 'Success' : 'Timeout')}
                      </Badge>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

