import type { Device } from '@/gen/v1/device_pb'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useDevicesContext } from '../devices-provider'
import type { DevicesDialogType } from '../../types'
import {
  BarChart2,
  Code,
  Code2,
  Edit2,
  MoreVertical,
  Settings2,
  ShieldAlert,
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

  const handleAction = (action: DevicesDialogType) => {
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
      <div className='flex items-center gap-1.5 shrink-0'>
        {boardName && (
          <Badge
            variant='outline'
            className='text-[10px] font-mono uppercase px-1.5 py-0 bg-muted/30'
          >
            {boardName}
          </Badge>
        )}

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='ghost'
              size='icon'
              className='h-8 w-8 text-muted-foreground hover:text-foreground data-[state=open]:bg-muted'
            >
              <MoreVertical className='h-4 w-4' />
              <span className='sr-only'>Menu opsi router</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' className='w-56'>
            <DropdownMenuLabel className='text-[11px] font-normal text-muted-foreground'>
              Operasi & Integrasi
            </DropdownMenuLabel>
            <DropdownMenuItem onClick={() => handleAction('isolation')}>
              <ShieldAlert className='me-2 h-4 w-4 text-amber-500' />
              Profil Isolir Router
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('webhook-scripts')}>
              <Code className='me-2 h-4 w-4 text-blue-500' />
              Script Webhook Event
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('ping-analytics')}>
              <BarChart2 className='me-2 h-4 w-4 text-blue-500' />
              Ping Analytics
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('ping-settings')}>
              <Settings2 className='me-2 h-4 w-4 text-indigo-500' />
              Pengaturan Ping
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('terminal')}>
              <Code2 className='me-2 h-4 w-4 text-emerald-500' />
              SSH Terminal
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('test')}>
              <Zap className='me-2 h-4 w-4 text-amber-500' />
              Uji Koneksi
            </DropdownMenuItem>

            <DropdownMenuSeparator />

            <DropdownMenuItem onClick={() => handleAction('edit')}>
              <Edit2 className='me-2 h-4 w-4' />
              Edit Perangkat
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className='text-destructive focus:text-destructive'
              onClick={() => handleAction('delete')}
            >
              <Trash2 className='me-2 h-4 w-4' />
              Hapus Perangkat
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
