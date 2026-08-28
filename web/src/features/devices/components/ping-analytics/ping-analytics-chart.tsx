import { useMemo } from 'react'
import { Activity, Loader2 } from 'lucide-react'
import type { PingMetricPointData } from '@/gen/v1/device_pb'

interface PingAnalyticsChartProps {
  points: PingMetricPointData[]
  avgRtt: number
  isLoading: boolean
}

export function PingAnalyticsChart({
  points,
  avgRtt,
  isLoading,
}: PingAnalyticsChartProps) {
  const chartPoints = useMemo(() => {
    if (!points || points.length === 0) return []
    return points.map((p, idx) => ({
      x: idx,
      y: p.rttMs,
      loss: p.packetLoss,
      time: p.timestamp,
      status: p.status,
    }))
  }, [points])

  const maxVal = useMemo(() => {
    if (!chartPoints.length) return 100
    const m = Math.max(...chartPoints.map((p) => p.y))
    return m > 0 ? Math.ceil(m * 1.25) : 50
  }, [chartPoints])

  if (isLoading && points.length === 0) {
    return (
      <div className='h-64 flex flex-col items-center justify-center border rounded-lg bg-card/50'>
        <Loader2 className='h-6 w-6 animate-spin text-primary' />
        <span className='text-xs text-muted-foreground mt-2'>Memuat data metrik...</span>
      </div>
    )
  }

  if (chartPoints.length === 0) {
    return (
      <div className='h-64 flex flex-col items-center justify-center border rounded-lg text-muted-foreground text-xs gap-1.5 bg-card/30'>
        <Activity className='h-8 w-8 text-muted-foreground/40' />
        <span className='font-medium text-foreground'>Tidak ada data metrik ping pada rentang waktu ini.</span>
        <span className='text-[11px]'>Pastikan fitur ping metrics pada router telah diaktifkan di Pengaturan.</span>
      </div>
    )
  }

  const firstTimestamp = points[0]?.timestamp
    ? new Date(points[0].timestamp).toLocaleTimeString()
    : ''
  const lastTimestamp = points[points.length - 1]?.timestamp
    ? new Date(points[points.length - 1].timestamp).toLocaleTimeString()
    : ''

  const pointsCount = Math.max(chartPoints.length, 1)

  return (
    <div className='border rounded-lg p-4 bg-card space-y-4 shadow-2xs'>
      {/* SVG Line Chart */}
      <div className='relative h-60 w-full'>
        <svg
          className='w-full h-full overflow-visible'
          viewBox={`0 0 ${pointsCount} 100`}
          preserveAspectRatio='none'
        >
          <defs>
            <linearGradient id='pingAreaGradient' x1='0' y1='0' x2='0' y2='1'>
              <stop offset='0%' stopColor='var(--primary)' stopOpacity='0.25' />
              <stop offset='100%' stopColor='var(--primary)' stopOpacity='0.0' />
            </linearGradient>
          </defs>

          {/* Horizontal Grid Lines */}
          <line x1='0' y1='25' x2={pointsCount} y2='25' stroke='currentColor' strokeOpacity='0.08' />
          <line x1='0' y1='50' x2={pointsCount} y2='50' stroke='currentColor' strokeOpacity='0.08' />
          <line x1='0' y1='75' x2={pointsCount} y2='75' stroke='currentColor' strokeOpacity='0.08' />

          {/* Area under line */}
          <polygon
            points={`0,100 ${chartPoints
              .map((p) => `${p.x},${100 - (p.y / maxVal) * 100}`)
              .join(' ')} ${chartPoints.length - 1},100`}
            fill='url(#pingAreaGradient)'
          />

          {/* Latency Polyline */}
          <polyline
            fill='none'
            stroke='var(--primary)'
            strokeWidth='1.5'
            strokeLinejoin='round'
            strokeLinecap='round'
            points={chartPoints
              .map((p) => `${p.x},${100 - (p.y / maxVal) * 100}`)
              .join(' ')}
          />

          {/* Average Baseline Reference Line */}
          {avgRtt > 0 && (
            <line
              x1='0'
              y1={100 - (avgRtt / maxVal) * 100}
              x2={pointsCount}
              y2={100 - (avgRtt / maxVal) * 100}
              stroke='#10b981'
              strokeDasharray='4 4'
              strokeWidth='1'
            />
          )}
        </svg>

        {/* Floating Labels */}
        <div className='absolute top-1 left-2 text-[10px] font-mono text-muted-foreground bg-background/80 px-1 rounded'>
          Max: {maxVal} ms
        </div>
        {avgRtt > 0 && (
          <div
            className='absolute right-2 text-[10px] font-mono text-emerald-600 dark:text-emerald-400 bg-background/80 px-1 rounded'
            style={{
              top: `${Math.max(8, Math.min(85, 100 - (avgRtt / maxVal) * 100))}%`,
            }}
          >
            Avg: {avgRtt.toFixed(1)} ms
          </div>
        )}
      </div>

      {/* Time Boundary Axis */}
      <div className='flex items-center justify-between text-[11px] text-muted-foreground font-mono px-1 border-t pt-2'>
        <span>{firstTimestamp}</span>
        <span className='text-[10px] text-muted-foreground/80 font-sans'>
          {chartPoints.length} titik sampel
        </span>
        <span>{lastTimestamp}</span>
      </div>
    </div>
  )
}

