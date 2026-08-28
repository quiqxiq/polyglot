import type { Device } from '@/gen/v1/device_pb'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useDevicesContext } from '../devices-provider'
import {
  BarChart2,
  Code2,
  Edit2,
  Settings2,
  Trash2,
  Zap,
} from 'lucide-react'

interface DeviceCardHeaderProps {
  device: Device
  isOnline: boolean
  boardName: string
}

export function DeviceCardHeader({
  device,
  isOnline,
  boardName,
}: DeviceCardHeaderProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()

  const handleAction = (action: 'terminal' | 'test' | 'edit' | 'delete' | 'ping-analytics' | 'ping-settings') => {
    setCurrentRow(device)
    setOpen(action)
  }

  return (
    <header className='flex items-start justify-between gap-2 border-b pb-3'>
      <div className='flex items-center gap-2.5 min-w-0'>
        {/* Status Indicator Dot with Glow */}
        <div className='relative flex h-3 w-3 shrink-0 items-center justify-center'>
          <span
            className={`absolute inline-flex h-full w-full rounded-full transition-opacity ${
              isOnline
                ? 'bg-emerald-400 opacity-75 animate-ping'
                : 'opacity-0'
            }`}
          />
          <span
            className={`relative inline-flex h-2.5 w-2.5 rounded-full ${
              isOnline
                ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]'
                : 'bg-muted-foreground/40'
            }`}
            title={isOnline ? 'Online / Connected' : 'Offline / Disconnected'}
          />
        </div>

        {/* Name and Host details */}
        <div className='min-w-0 flex-1'>
          <h3
            className='font-semibold text-sm sm:text-base leading-tight truncate text-foreground'
            title={device.name}
          >
            {device.name}
          </h3>
          <p className='text-xs text-muted-foreground font-mono mt-0.5 truncate flex items-center gap-1.5'>
            <span>{device.host}:{device.port}</span>
            {device.sshPort ? (
              <span className='text-[10px] text-muted-foreground/70'>
                (SSH {device.sshPort})
              </span>
            ) : null}
          </p>
        </div>
      </div>

      {/* Badges & Action Toolbar */}
      <div className='flex items-center gap-0.5 sm:gap-1 shrink-0'>
        {boardName && (
          <Badge
            variant='outline'
            className='text-[10px] font-mono uppercase hidden sm:inline-flex px-1.5 py-0 bg-muted/30 mr-0.5'
          >
            {boardName}
          </Badge>
        )}

        <TooltipProvider delayDuration={200}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-blue-500 hover:bg-blue-500/10 transition-colors'
                onClick={() => handleAction('ping-analytics')}
              >
                <BarChart2 className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Analisis Ping Historis
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-indigo-500 hover:bg-indigo-500/10 transition-colors'
                onClick={() => handleAction('ping-settings')}
              >
                <Settings2 className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Pengaturan Ping
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-emerald-500 hover:bg-emerald-500/10 transition-colors'
                onClick={() => handleAction('terminal')}
              >
                <Code2 className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Terminal SSH Web
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-amber-500 hover:bg-amber-500/10 transition-colors'
                onClick={() => handleAction('test')}
              >
                <Zap className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Uji Koneksi
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors'
                onClick={() => handleAction('edit')}
              >
                <Edit2 className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Edit Perangkat
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors'
                onClick={() => handleAction('delete')}
              >
                <Trash2 className='h-3.5 w-3.5' />
              </Button>
            </TooltipTrigger>
            <TooltipContent side='bottom' className='text-xs'>
              Hapus Perangkat
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </header>
  )
}

