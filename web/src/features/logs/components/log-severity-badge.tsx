import { Badge } from '@/components/ui/badge'
import { AlertCircle, AlertTriangle, Info, Bug } from 'lucide-react'
import { LogSeverity } from '../types'

interface LogSeverityBadgeProps {
  severity: LogSeverity
  className?: string
}

export function LogSeverityBadge({ severity, className = '' }: LogSeverityBadgeProps) {
  if (severity === 'error') {
    return (
      <Badge
        variant='outline'
        className={`h-5 gap-1 border-red-500/40 bg-red-500/10 text-red-500 dark:text-red-400 font-mono text-[10px] font-semibold select-none ${className}`}
      >
        <AlertCircle className='size-2.5 shrink-0' />
        ERROR
      </Badge>
    )
  }

  if (severity === 'warning') {
    return (
      <Badge
        variant='outline'
        className={`h-5 gap-1 border-amber-500/40 bg-amber-500/10 text-amber-600 dark:text-amber-400 font-mono text-[10px] font-semibold select-none ${className}`}
      >
        <AlertTriangle className='size-2.5 shrink-0' />
        WARN
      </Badge>
    )
  }

  if (severity === 'debug') {
    return (
      <Badge
        variant='secondary'
        className={`h-5 gap-1 font-mono text-[10px] font-normal text-muted-foreground select-none ${className}`}
      >
        <Bug className='size-2.5 shrink-0' />
        DEBUG
      </Badge>
    )
  }

  return (
    <Badge
      variant='outline'
      className={`h-5 gap-1 border-border/60 bg-muted/40 text-muted-foreground font-mono text-[10px] font-normal select-none ${className}`}
    >
      <Info className='size-2.5 shrink-0 text-muted-foreground' />
      INFO
    </Badge>
  )
}
