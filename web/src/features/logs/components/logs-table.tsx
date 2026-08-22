import { LogItem } from '../types'
import { LogEntryRow } from './log-entry-row'
import { Terminal, Table as TableIcon, SearchX, Inbox } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useSmartScroll } from '../hooks/use-smart-scroll'
import { JumpToLatestButton } from './jump-to-latest-button'

interface LogsTableProps {
  logs: LogItem[]
  highlight?: string
  isAutoScroll?: boolean
  isLoading?: boolean
  viewMode?: 'structured' | 'raw'
  onViewModeChange?: (mode: 'structured' | 'raw') => void
}

export function LogsTable({
  logs,
  highlight = '',
  isAutoScroll = true,
  isLoading = false,
  viewMode = 'raw',
  onViewModeChange,
}: LogsTableProps) {
  const { containerRef, handleScroll, scrollToBottom, showScrollBottomBtn } =
    useSmartScroll({
      dependencies: [logs],
      isAutoScrollEnabled: isAutoScroll,
      threshold: 50,
    })

  return (
    <div className='relative flex flex-1 min-h-0 flex-col rounded-lg border bg-card shadow-xs overflow-hidden'>
      {/* Table Header Bar with View Switcher (Consistent with StructuredLogTable) */}
      <div className='shrink-0 flex items-center justify-between border-b bg-muted/60 px-3 py-1.5 text-xs text-muted-foreground select-none'>
        <div className='flex items-center gap-2 font-medium text-foreground'>
          {/* macOS Terminal Traffic Light Dots */}
          <svg
            width='36'
            height='10'
            viewBox='0 0 36 10'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            className='shrink-0'
            aria-hidden='true'
          >
            <circle cx='5' cy='5' r='4.5' fill='#ff5f56' />
            <circle cx='18' cy='5' r='4.5' fill='#ffbd2e' />
            <circle cx='31' cy='5' r='4.5' fill='#27c93f' />
          </svg>
          <div className='flex items-center gap-1.5 font-mono font-medium text-xs text-foreground'>
            <Terminal className='size-3.5 text-primary' />
            <span>Raw Log Stream</span>
          </div>
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

      {/* Column Headers Sub-Header */}
      <div className='shrink-0 flex items-center justify-between border-b bg-muted px-4 py-2 text-xs font-mono font-semibold text-foreground select-none shadow-xs'>
        <div className='flex items-center gap-3 text-xs font-mono font-semibold'>
          <span className='w-20 text-muted-foreground'>TIME</span>
          <span className='w-16 text-muted-foreground'>SEVERITY</span>
          <span className='w-32 text-muted-foreground'>TOPICS</span>
          <span className='text-muted-foreground hidden sm:inline'>MESSAGE</span>
        </div>
      </div>

      {/* Log entries scroll container */}
      <div
        ref={containerRef}
        onScroll={handleScroll}
        style={{ overflowAnchor: 'none' }}
        className='flex-1 min-h-0 overflow-y-auto divide-y divide-border/40 font-mono pb-20'
      >
        {logs.length === 0 ? (
          <div className='flex flex-col items-center justify-center h-full min-h-[260px] text-center text-muted-foreground p-8'>
            {isLoading ? (
              <div className='flex flex-col items-center gap-3'>
                <div className='size-6 animate-spin rounded-full border-2 border-primary border-t-transparent' />
                <p className='text-xs font-sans'>Connecting to RouterOS live stream...</p>
              </div>
            ) : highlight ? (
              <div className='flex flex-col items-center gap-2'>
                <SearchX className='size-8 text-muted-foreground' />
                <p className='text-sm font-medium text-foreground'>No logs match your search filter</p>
                <p className='text-xs text-muted-foreground'>Try adjusting your search query or severity filter</p>
              </div>
            ) : (
              <div className='flex flex-col items-center gap-2'>
                <Inbox className='size-8 text-muted-foreground' />
                <p className='text-sm font-medium text-foreground'>No logs received yet</p>
                <p className='text-xs text-muted-foreground'>Waiting for RouterOS log events to appear...</p>
              </div>
            )}
          </div>
        ) : (
          <>
            {logs.map((log) => (
              <LogEntryRow key={log.id} log={log} highlight={highlight} />
            ))}
            <div className='h-12' />
          </>
        )}
      </div>

      {/* Reusable Floating Jump to Latest Button (anchored bottom-right) */}
      <JumpToLatestButton
        visible={showScrollBottomBtn && logs.length > 0}
        onClick={() => scrollToBottom(true)}
      />
    </div>
  )
}
