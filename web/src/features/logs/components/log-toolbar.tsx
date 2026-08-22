import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Search,
  X,
  Play,
  Pause,
  Trash2,
  Download,
  AlertCircle,
  AlertTriangle,
  Info,
  Radio,
} from 'lucide-react'
import { SeverityFilter } from '../types'

interface LogToolbarProps {
  searchTerm: string
  onSearchChange: (value: string) => void
  severityFilter: SeverityFilter
  onSeverityChange: (severity: SeverityFilter) => void
  isPaused: boolean
  onTogglePause: () => void
  isAutoScroll: boolean
  onToggleAutoScroll: () => void
  onClear: () => void
  onExportTxt: () => void
  onExportJson: () => void
  totalCount: number
  filteredCount: number
  isStreaming: boolean
}

export function LogToolbar({
  searchTerm,
  onSearchChange,
  severityFilter,
  onSeverityChange,
  isPaused,
  onTogglePause,
  isAutoScroll,
  onToggleAutoScroll,
  onClear,
  onExportTxt,
  onExportJson,
  totalCount,
  filteredCount,
  isStreaming,
}: LogToolbarProps) {
  return (
    <div className='flex flex-wrap items-center justify-between gap-2.5 rounded-lg border bg-card p-2 shadow-xs'>
      {/* Left side: Search & Quick Severity Filters */}
      <div className='flex flex-1 flex-wrap items-center gap-2 min-w-[260px]'>
        {/* Search Input */}
        <div className='relative flex-1 min-w-[180px] max-w-sm'>
          <Search className='absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground' />
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder='Filter logs by text, topic, or time...'
            className='h-8 pl-8 pr-8 text-xs'
          />
          {searchTerm && (
            <button
              onClick={() => onSearchChange('')}
              className='absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
              title='Clear search'
            >
              <X className='size-3.5' />
            </button>
          )}
        </div>

        {/* Severity Filter Pills */}
        <div className='flex items-center gap-1'>
          <Button
            variant={severityFilter === 'all' ? 'secondary' : 'ghost'}
            size='sm'
            onClick={() => onSeverityChange('all')}
            className='h-8 px-2.5 text-xs font-medium'
          >
            All
          </Button>
          <Button
            variant={severityFilter === 'error' ? 'destructive' : 'ghost'}
            size='sm'
            onClick={() => onSeverityChange(severityFilter === 'error' ? 'all' : 'error')}
            className={`h-8 gap-1 px-2.5 text-xs font-medium ${
              severityFilter !== 'error' ? 'text-destructive hover:bg-destructive/10' : ''
            }`}
          >
            <AlertCircle className='size-3' />
            <span>Error</span>
          </Button>
          <Button
            variant={severityFilter === 'warning' ? 'default' : 'ghost'}
            size='sm'
            onClick={() => onSeverityChange(severityFilter === 'warning' ? 'all' : 'warning')}
            className={`h-8 gap-1 px-2.5 text-xs font-medium ${
              severityFilter === 'warning'
                ? 'bg-amber-600 hover:bg-amber-700 text-white'
                : 'text-amber-600 hover:bg-amber-500/10'
            }`}
          >
            <AlertTriangle className='size-3' />
            <span>Warning</span>
          </Button>
          <Button
            variant={severityFilter === 'info' ? 'secondary' : 'ghost'}
            size='sm'
            onClick={() => onSeverityChange(severityFilter === 'info' ? 'all' : 'info')}
            className='h-8 gap-1 px-2.5 text-xs font-medium text-muted-foreground'
          >
            <Info className='size-3' />
            <span>Info</span>
          </Button>
        </div>
      </div>

      {/* Right side: Stream Controls, Auto Scroll & Export */}
      <div className='flex items-center gap-2 shrink-0'>
        {/* Stream Status Indicator */}
        <div className='hidden sm:flex items-center gap-1.5 px-2 py-1 text-xs text-muted-foreground font-mono'>
          {isPaused ? (
            <span className='flex items-center gap-1 text-amber-500 font-medium'>
              <span className='size-2 rounded-full bg-amber-500' />
              PAUSED
            </span>
          ) : isStreaming ? (
            <span className='flex items-center gap-1 text-emerald-500 font-medium'>
              <span className='relative flex size-2'>
                <span className='absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75' />
                <span className='relative inline-flex size-2 rounded-full bg-emerald-500' />
              </span>
              LIVE
            </span>
          ) : (
            <span className='flex items-center gap-1 text-muted-foreground'>
              <Radio className='size-3 animate-pulse' />
              CONNECTING
            </span>
          )}
        </div>

        {/* Pause/Resume button */}
        <Button
          variant='outline'
          size='sm'
          onClick={onTogglePause}
          className={`h-8 gap-1 text-xs font-medium ${
            isPaused
              ? 'border-amber-500/40 bg-amber-500/10 text-amber-600 hover:bg-amber-500/20'
              : 'text-muted-foreground hover:text-foreground'
          }`}
          title={isPaused ? 'Resume live streaming' : 'Pause streaming to inspect'}
        >
          {isPaused ? (
            <>
              <Play className='size-3.5 fill-current' />
              <span>Resume</span>
            </>
          ) : (
            <>
              <Pause className='size-3.5 fill-current' />
              <span>Pause</span>
            </>
          )}
        </Button>

        {/* Auto Scroll toggle */}
        <div className='flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs'>
          <Switch
            id='autoscroll'
            checked={isAutoScroll}
            onCheckedChange={onToggleAutoScroll}
            className='scale-75'
          />
          <Label htmlFor='autoscroll' className='cursor-pointer text-xs text-muted-foreground select-none'>
            Auto-scroll
          </Label>
        </div>

        {/* Clear buffer button */}
        <Button
          variant='outline'
          size='sm'
          onClick={onClear}
          className='h-8 gap-1 text-xs text-muted-foreground hover:text-destructive'
          title='Clear log buffer'
        >
          <Trash2 className='size-3.5' />
          <span className='hidden md:inline'>Clear</span>
        </Button>

        {/* Export dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='outline' size='sm' className='h-8 gap-1 text-xs'>
              <Download className='size-3.5' />
              <span className='hidden md:inline'>Export</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' className='text-xs'>
            <DropdownMenuItem onClick={onExportTxt} className='gap-2 text-xs'>
              <span>Export as Text (.txt)</span>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={onExportJson} className='gap-2 text-xs'>
              <span>Export as JSON (.json)</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Counter Pill */}
        <div className='hidden lg:flex items-center text-[11px] text-muted-foreground font-mono bg-muted/60 px-2 py-1 rounded-md border'>
          {filteredCount === totalCount ? (
            <span>{totalCount} logs</span>
          ) : (
            <span>
              {filteredCount} / {totalCount} logs
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
