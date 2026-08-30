import {
  Activity,
  ArrowDown,
  ArrowUp,
  Clock,
  TrendingDown,
  CheckCircle2,
} from 'lucide-react'
import type { PingSummaryStats } from '../../types'

interface PingAnalyticsKpiGridProps {
  summary: PingSummaryStats
}

export function PingAnalyticsKpiGrid({ summary }: PingAnalyticsKpiGridProps) {
  const isLoss = summary.packetLossPct > 0

  return (
    <div className='grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3'>
      {/* Min Latency */}
      <div className='relative overflow-hidden rounded-xl border bg-card/60 backdrop-blur-xs p-3 shadow-2xs flex flex-col justify-between min-w-0 transition-all hover:border-emerald-500/30'>
        <div className='flex items-center justify-between gap-1'>
          <span className='text-[10px] sm:text-[11px] font-semibold text-muted-foreground uppercase tracking-wider truncate'>
            Min RTT
          </span>
          <div className='flex h-5 w-5 sm:h-6 sm:w-6 shrink-0 items-center justify-center rounded-md bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'>
            <ArrowDown className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
          </div>
        </div>
        <div className='mt-2 flex items-baseline gap-1 min-w-0'>
          <span className='text-lg sm:text-xl font-bold font-mono tracking-tight text-emerald-600 dark:text-emerald-400 truncate'>
            {summary.minRtt > 0 ? summary.minRtt.toFixed(1) : '-'}
          </span>
          {summary.minRtt > 0 && (
            <span className='text-xs font-normal text-muted-foreground'>ms</span>
          )}
        </div>
      </div>

      {/* Avg Latency */}
      <div className='relative overflow-hidden rounded-xl border bg-card/60 backdrop-blur-xs p-3 shadow-2xs flex flex-col justify-between min-w-0 transition-all hover:border-primary/30'>
        <div className='flex items-center justify-between gap-1'>
          <span className='text-[10px] sm:text-[11px] font-semibold text-muted-foreground uppercase tracking-wider truncate'>
            Avg RTT
          </span>
          <div className='flex h-5 w-5 sm:h-6 sm:w-6 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary'>
            <Activity className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
          </div>
        </div>
        <div className='mt-2 flex items-baseline gap-1 min-w-0'>
          <span className='text-lg sm:text-xl font-bold font-mono tracking-tight text-primary truncate'>
            {summary.avgRtt > 0 ? summary.avgRtt.toFixed(1) : '-'}
          </span>
          {summary.avgRtt > 0 && (
            <span className='text-xs font-normal text-muted-foreground'>ms</span>
          )}
        </div>
      </div>

      {/* Max Latency */}
      <div className='relative overflow-hidden rounded-xl border bg-card/60 backdrop-blur-xs p-3 shadow-2xs flex flex-col justify-between min-w-0 transition-all hover:border-amber-500/30'>
        <div className='flex items-center justify-between gap-1'>
          <span className='text-[10px] sm:text-[11px] font-semibold text-muted-foreground uppercase tracking-wider truncate'>
            Max RTT
          </span>
          <div className='flex h-5 w-5 sm:h-6 sm:w-6 shrink-0 items-center justify-center rounded-md bg-amber-500/10 text-amber-600 dark:text-amber-400'>
            <ArrowUp className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
          </div>
        </div>
        <div className='mt-2 flex items-baseline gap-1 min-w-0'>
          <span className='text-lg sm:text-xl font-bold font-mono tracking-tight text-amber-600 dark:text-amber-400 truncate'>
            {summary.maxRtt > 0 ? summary.maxRtt.toFixed(1) : '-'}
          </span>
          {summary.maxRtt > 0 && (
            <span className='text-xs font-normal text-muted-foreground'>ms</span>
          )}
        </div>
      </div>

      {/* Packet Loss */}
      <div
        className={`relative overflow-hidden rounded-xl border bg-card/60 backdrop-blur-xs p-3 shadow-2xs flex flex-col justify-between min-w-0 transition-all ${
          isLoss
            ? 'border-destructive/30 hover:border-destructive/50'
            : 'hover:border-emerald-500/30'
        }`}
      >
        <div className='flex items-center justify-between gap-1'>
          <span className='text-[10px] sm:text-[11px] font-semibold text-muted-foreground uppercase tracking-wider truncate'>
            Packet Loss
          </span>
          <div
            className={`flex h-5 w-5 sm:h-6 sm:w-6 shrink-0 items-center justify-center rounded-md ${
              isLoss
                ? 'bg-destructive/10 text-destructive'
                : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
            }`}
          >
            {isLoss ? (
              <TrendingDown className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
            ) : (
              <CheckCircle2 className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
            )}
          </div>
        </div>
        <div className='mt-2 flex items-baseline gap-1 min-w-0'>
          <span
            className={`text-lg sm:text-xl font-bold font-mono tracking-tight truncate ${
              isLoss
                ? 'text-destructive'
                : 'text-emerald-600 dark:text-emerald-400'
            }`}
          >
            {summary.packetLossPct.toFixed(1)}
          </span>
          <span className='text-xs font-normal text-muted-foreground'>%</span>
        </div>
      </div>

      {/* Total Samples */}
      <div className='relative overflow-hidden rounded-xl border bg-card/60 backdrop-blur-xs p-3 shadow-2xs flex flex-col justify-between min-w-0 col-span-2 sm:col-span-1 transition-all hover:border-primary/30'>
        <div className='flex items-center justify-between gap-1'>
          <span className='text-[10px] sm:text-[11px] font-semibold text-muted-foreground uppercase tracking-wider truncate'>
            Total Paket
          </span>
          <div className='flex h-5 w-5 sm:h-6 sm:w-6 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground'>
            <Clock className='h-3 w-3 sm:h-3.5 sm:w-3.5' />
          </div>
        </div>
        <div className='mt-2 flex items-baseline gap-1 min-w-0'>
          <span className='text-lg sm:text-xl font-bold font-mono tracking-tight text-foreground truncate'>
            {summary.totalSamples.toLocaleString()}
          </span>
          <span className='text-xs font-normal text-muted-foreground'>pkt</span>
        </div>
      </div>
    </div>
  )
}
