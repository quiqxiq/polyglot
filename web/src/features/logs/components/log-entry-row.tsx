import { LogItem } from '../types'
import { LogSeverityBadge } from './log-severity-badge'
import { LogTopicsBadge } from './log-topics-badge'
import { LogHighlightText } from './log-highlight-text'

interface LogEntryRowProps {
  log: LogItem
  highlight?: string
}

export function LogEntryRow({ log, highlight = '' }: LogEntryRowProps) {
  let rowBorderColor = 'border-l-transparent hover:bg-muted/40'

  if (log.severity === 'error') {
    rowBorderColor = 'border-l-red-500 bg-red-500/5 hover:bg-red-500/10'
  } else if (log.severity === 'warning') {
    rowBorderColor = 'border-l-amber-500 bg-amber-500/5 hover:bg-amber-500/10'
  }

  return (
    <div
      className={`group flex items-start gap-2 border-l-2 py-1.5 px-3 text-xs font-mono transition-colors ${rowBorderColor}`}
    >
      {/* Time */}
      <span className='shrink-0 text-muted-foreground w-20 select-none text-[11px] pt-0.5'>
        {log.time}
      </span>

      {/* Severity */}
      <div className='shrink-0 pt-0.5 w-16'>
        <LogSeverityBadge severity={log.severity} />
      </div>

      {/* Topics */}
      <div className='shrink-0 max-w-[140px] pt-0.5'>
        <LogTopicsBadge topics={log.topics} />
      </div>

      {/* Message */}
      <div
        className={`flex-1 break-all leading-relaxed ${
          log.severity === 'error'
            ? 'text-red-400 font-medium'
            : log.severity === 'warning'
            ? 'text-amber-300'
            : 'text-foreground/90'
        }`}
      >
        <LogHighlightText text={log.message} highlight={highlight} />
      </div>
    </div>
  )
}
