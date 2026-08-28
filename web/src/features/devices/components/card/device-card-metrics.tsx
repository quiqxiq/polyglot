import { Cpu, HardDrive, Clock, Layers } from 'lucide-react'

interface DeviceCardMetricsProps {
  cpuUsage: number
  memUsage: number
  uptime: string
  version: string
}

function getUsageColor(pct: number): string {
  if (pct >= 90) return 'bg-rose-500'
  if (pct >= 70) return 'bg-amber-500'
  return 'bg-emerald-500'
}

function getUsageTextColor(pct: number): string {
  if (pct >= 90) return 'text-rose-600 dark:text-rose-400'
  if (pct >= 70) return 'text-amber-600 dark:text-amber-400'
  return 'text-foreground'
}

export function DeviceCardMetrics({
  cpuUsage,
  memUsage,
  uptime,
  version,
}: DeviceCardMetricsProps) {
  return (
    <section className='grid grid-cols-2 gap-3 py-3 border-b text-xs'>
      {/* CPU Usage Bar */}
      <div>
        <div className='flex items-center justify-between mb-1 text-muted-foreground'>
          <span className='flex items-center gap-1 text-[11px] font-medium'>
            <Cpu className='h-3 w-3 text-muted-foreground/80' />
            CPU
          </span>
          <span className={`font-mono font-semibold ${getUsageTextColor(cpuUsage)}`}>
            {cpuUsage}%
          </span>
        </div>
        <div className='h-1.5 w-full bg-secondary/80 rounded-full overflow-hidden'>
          <div
            className={`h-full transition-all duration-500 rounded-full ${getUsageColor(
              cpuUsage
            )}`}
            style={{ width: `${Math.min(100, Math.max(0, cpuUsage))}%` }}
          />
        </div>
      </div>

      {/* Memory Usage Bar */}
      <div>
        <div className='flex items-center justify-between mb-1 text-muted-foreground'>
          <span className='flex items-center gap-1 text-[11px] font-medium'>
            <HardDrive className='h-3 w-3 text-muted-foreground/80' />
            Memory
          </span>
          <span className={`font-mono font-semibold ${getUsageTextColor(memUsage)}`}>
            {memUsage}%
          </span>
        </div>
        <div className='h-1.5 w-full bg-secondary/80 rounded-full overflow-hidden'>
          <div
            className={`h-full transition-all duration-500 rounded-full ${getUsageColor(
              memUsage
            )}`}
            style={{ width: `${Math.min(100, Math.max(0, memUsage))}%` }}
          />
        </div>
      </div>

      {/* Uptime and Version */}
      <div className='flex items-center justify-between col-span-1 text-[11px] text-muted-foreground pt-1'>
        <span className='flex items-center gap-1'>
          <Clock className='h-3 w-3 text-muted-foreground/70' />
          Uptime:
        </span>
        <span className='font-mono text-foreground font-medium truncate max-w-[90px]' title={uptime}>
          {uptime}
        </span>
      </div>

      <div className='flex items-center justify-between col-span-1 text-[11px] text-muted-foreground pt-1'>
        <span className='flex items-center gap-1'>
          <Layers className='h-3 w-3 text-muted-foreground/70' />
          OS:
        </span>
        <span className='font-mono text-foreground font-medium truncate max-w-[90px]' title={version}>
          {version}
        </span>
      </div>
    </section>
  )
}

