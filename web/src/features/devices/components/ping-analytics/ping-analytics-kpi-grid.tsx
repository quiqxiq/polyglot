import { ArrowDown, ArrowUp, Clock, TrendingDown, Wifi } from 'lucide-react'
import type { PingSummaryStats } from '../../types'

interface PingAnalyticsKpiGridProps {
  summary: PingSummaryStats
}

export function PingAnalyticsKpiGrid({ summary }: PingAnalyticsKpiGridProps) {
  return (
    <div className='grid grid-cols-2 sm:grid-cols-5 gap-3'>
      {/* Min Latency */}
      <div className='rounded-lg border p-3 bg-card shadow-2xs'>
        <span className='text-[11px] text-muted-foreground font-medium'>Min Latency</span>
        <div className='text-lg font-bold font-mono text-emerald-600 dark:text-emerald-400 flex items-center gap-1 mt-0.5'>
          <ArrowDown className='h-3.5 w-3.5 shrink-0' />
          <span>{summary.minRtt.toFixed(1)} ms</span>
        </div>
      </div>

      {/* Avg Latency */}
      <div className='rounded-lg border p-3 bg-card shadow-2xs'>
        <span className='text-[11px] text-muted-foreground font-medium'>Avg Latency</span>
        <div className='text-lg font-bold font-mono text-primary flex items-center gap-1 mt-0.5'>
          <Wifi className='h-3.5 w-3.5 shrink-0' />
          <span>{summary.avgRtt.toFixed(1)} ms</span>
        </div>
      </div>

      {/* Max Latency */}
      <div className='rounded-lg border p-3 bg-card shadow-2xs'>
        <span className='text-[11px] text-muted-foreground font-medium'>Max Latency</span>
        <div className='text-lg font-bold font-mono text-amber-600 dark:text-amber-400 flex items-center gap-1 mt-0.5'>
          <ArrowUp className='h-3.5 w-3.5 shrink-0' />
          <span>{summary.maxRtt.toFixed(1)} ms</span>
        </div>
      </div>

      {/* Packet Loss */}
      <div className='rounded-lg border p-3 bg-card shadow-2xs'>
        <span className='text-[11px] text-muted-foreground font-medium'>Packet Loss</span>
        <div
          className={`text-lg font-bold font-mono flex items-center gap-1 mt-0.5 ${
            summary.packetLossPct > 0
              ? 'text-destructive'
              : 'text-emerald-600 dark:text-emerald-400'
          }`}
        >
          <TrendingDown className='h-3.5 w-3.5 shrink-0' />
          <span>{summary.packetLossPct.toFixed(1)}%</span>
        </div>
      </div>

      {/* Total Samples */}
      <div className='rounded-lg border p-3 bg-card shadow-2xs col-span-2 sm:col-span-1'>
        <span className='text-[11px] text-muted-foreground font-medium'>Total Samples</span>
        <div className='text-lg font-bold font-mono text-foreground flex items-center gap-1 mt-0.5'>
          <Clock className='h-3.5 w-3.5 text-muted-foreground shrink-0' />
          <span>{summary.totalSamples.toLocaleString()}</span>
        </div>
      </div>
    </div>
  )
}

