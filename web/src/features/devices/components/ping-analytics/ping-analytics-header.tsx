import {
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Radio,
  RefreshCw,
} from 'lucide-react'

interface PingAnalyticsHeaderProps {
  deviceName: string
  host: string
  pingTarget: string
  isStreaming: boolean
  timescaledbAvailable: boolean
  isFetching: boolean
  onRefresh: () => void
}

export function PingAnalyticsHeader({
  deviceName,
  host,
  pingTarget,
  isStreaming,
  timescaledbAvailable,
  isFetching,
  onRefresh,
}: PingAnalyticsHeaderProps) {
  return (
    <DialogHeader className='p-6 pb-4 border-b'>
      <div className='flex items-center justify-between gap-4'>
        <div className='flex items-center gap-3 min-w-0'>
          <div className='rounded-lg bg-primary/10 p-2 text-primary shrink-0'>
            <Activity className='h-5 w-5' />
          </div>
          <div className='min-w-0'>
            <DialogTitle className='text-base font-semibold flex items-center gap-2 flex-wrap'>
              <span className='truncate'>Analisis Ping Streaming — {deviceName}</span>
              {isStreaming && (
                <Badge
                  variant='outline'
                  className='text-[10px] bg-emerald-500/10 text-emerald-600 border-emerald-300 dark:border-emerald-700/50 gap-1 animate-pulse'
                >
                  <Radio className='h-2.5 w-2.5' />
                  Live Stream
                </Badge>
              )}
            </DialogTitle>
            <DialogDescription className='text-xs text-muted-foreground flex items-center gap-2 mt-0.5'>
              <span>
                Target: <code className='font-mono text-primary font-semibold'>{pingTarget}</code>
              </span>
              <span>•</span>
              <span>
                Host: <code className='font-mono'>{host}</code>
              </span>
            </DialogDescription>
          </div>
        </div>

        <div className='flex items-center gap-2 shrink-0'>
          {timescaledbAvailable ? (
            <Badge
              variant='outline'
              className='bg-emerald-500/10 text-emerald-600 border-emerald-200 dark:border-emerald-800 text-xs gap-1 hidden sm:inline-flex'
            >
              <CheckCircle2 className='h-3 w-3' />
              TimescaleDB Aktif
            </Badge>
          ) : (
            <Badge variant='destructive' className='text-xs gap-1 hidden sm:inline-flex'>
              <AlertTriangle className='h-3 w-3' />
              TimescaleDB Nonaktif
            </Badge>
          )}
          <Button
            variant='outline'
            size='icon'
            className='h-8 w-8'
            onClick={onRefresh}
            disabled={isFetching}
            title='Muat Ulang Data Historis'
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isFetching ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>
    </DialogHeader>
  )
}

