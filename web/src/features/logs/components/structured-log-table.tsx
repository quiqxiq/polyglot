import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Clock, SearchX, Inbox, Table as TableIcon, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { LogItem } from '../types'
import { LogSeverityBadge } from './log-severity-badge'
import { LogTopicsBadge } from './log-topics-badge'
import { LogHighlightText } from './log-highlight-text'
import { JumpToLatestButton } from './jump-to-latest-button'
import { useSmartScroll } from '../hooks/use-smart-scroll'

interface StructuredLogTableProps {
  logs: LogItem[]
  highlight?: string
  isAutoScroll?: boolean
  isLoading?: boolean
  viewMode?: 'structured' | 'raw'
  onViewModeChange?: (mode: 'structured' | 'raw') => void
}

export function StructuredLogTable({
  logs,
  highlight = '',
  isAutoScroll = true,
  isLoading = false,
  viewMode = 'structured',
  onViewModeChange,
}: StructuredLogTableProps) {
  const { containerRef, handleScroll, scrollToBottom, showScrollBottomBtn } =
    useSmartScroll({
      dependencies: [logs],
      isAutoScrollEnabled: isAutoScroll,
      threshold: 50,
    })

  return (
    <div className='relative flex flex-1 min-h-0 flex-col rounded-lg border bg-card shadow-xs overflow-hidden'>
      {/* Table Header Bar with View Switcher */}
      <div className='shrink-0 flex items-center justify-between border-b bg-muted/60 px-3 py-1.5 text-xs text-muted-foreground select-none'>
        <div className='flex items-center gap-2 font-medium text-foreground'>
          <TableIcon className='size-3.5 text-primary' />
          <span className='font-mono font-medium text-xs'>Structured Log View</span>
        </div>

        {onViewModeChange && (
          <div className='flex items-center rounded-md border bg-background/80 p-0.5 text-xs'>
            <Button
              variant='ghost'
              size='sm'
              onClick={() => onViewModeChange('structured')}
              className={`h-6 gap-1 px-2 text-[11px] font-medium transition-all ${
                viewMode === 'structured'
                  ? 'bg-muted text-foreground font-semibold shadow-xs'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              }`}
              title='Structured Table View'
            >
              <TableIcon className='size-3' />
              <span>Table</span>
            </Button>
            <Button
              variant='ghost'
              size='sm'
              onClick={() => onViewModeChange('raw')}
              className={`h-6 gap-1 px-2 text-[11px] font-medium transition-all ${
                viewMode === 'raw'
                  ? 'bg-muted text-foreground font-semibold shadow-xs'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              }`}
              title='Switch to Raw Terminal View'
            >
              <Terminal className='size-3' />
              <span>Raw</span>
            </Button>
          </div>
        )}
      </div>

      {/* Scrollable table container */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        style={{ overflowAnchor: 'none' }}
        className='flex-1 min-h-0 overflow-y-auto overflow-x-auto pb-24'
      >
        <table className='w-full caption-bottom text-sm'>
          <TableHeader className='sticky top-0 z-20 bg-muted border-b shadow-xs'>
            <TableRow className='hover:bg-muted border-b select-none bg-muted'>
              <TableHead className='sticky top-0 z-20 bg-muted w-[100px] text-xs font-mono font-semibold'>
                TIME
              </TableHead>
              <TableHead className='sticky top-0 z-20 bg-muted w-[90px] text-xs font-mono font-semibold'>
                SEVERITY
              </TableHead>
              <TableHead className='sticky top-0 z-20 bg-muted w-[150px] text-xs font-mono font-semibold'>
                TOPICS
              </TableHead>
              <TableHead className='sticky top-0 z-20 bg-muted min-w-[280px] text-xs font-mono font-semibold'>
                MESSAGE
              </TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className='h-52 text-center text-muted-foreground'>
                  {isLoading ? (
                    <div className='flex flex-col items-center justify-center gap-2'>
                      <div className='size-5 animate-spin rounded-full border-2 border-primary border-t-transparent' />
                      <span className='text-xs'>Connecting to RouterOS log stream...</span>
                    </div>
                  ) : highlight ? (
                    <div className='flex flex-col items-center justify-center gap-1.5'>
                      <SearchX className='size-6 text-muted-foreground' />
                      <span className='text-xs font-medium'>No logs match your search filter</span>
                    </div>
                  ) : (
                    <div className='flex flex-col items-center justify-center gap-1.5'>
                      <Inbox className='size-6 text-muted-foreground' />
                      <span className='text-xs'>No logs received yet</span>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => {
                let rowBg = 'hover:bg-muted/50'
                if (log.severity === 'error') {
                  rowBg = 'bg-red-500/5 hover:bg-red-500/10'
                } else if (log.severity === 'warning') {
                  rowBg = 'bg-amber-500/5 hover:bg-amber-500/10'
                }

                return (
                  <TableRow key={log.id} className={`font-mono text-xs transition-colors ${rowBg}`}>
                    {/* Time */}
                    <TableCell className='text-muted-foreground text-[11px] whitespace-nowrap py-2'>
                      <div className='flex items-center gap-1'>
                        <Clock className='size-3 shrink-0 text-muted-foreground hidden sm:inline' />
                        <span>{log.time}</span>
                      </div>
                    </TableCell>

                    {/* Severity */}
                    <TableCell className='py-2'>
                      <LogSeverityBadge severity={log.severity} />
                    </TableCell>

                    {/* Topics */}
                    <TableCell className='py-2'>
                      <LogTopicsBadge topics={log.topics} />
                    </TableCell>

                    {/* Message */}
                    <TableCell
                      className={`text-xs py-2 break-all leading-relaxed ${
                        log.severity === 'error'
                          ? 'text-red-500 dark:text-red-400 font-medium'
                          : log.severity === 'warning'
                          ? 'text-amber-600 dark:text-amber-400'
                          : 'text-foreground'
                      }`}
                    >
                      <LogHighlightText text={log.message} highlight={highlight} />
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </table>
        <div className='h-16' />
      </div>

      {/* Floating Jump to Latest Button anchored bottom-right */}
      <JumpToLatestButton
        visible={showScrollBottomBtn && logs.length > 0}
        onClick={() => scrollToBottom(true)}
      />
    </div>
  )
}
