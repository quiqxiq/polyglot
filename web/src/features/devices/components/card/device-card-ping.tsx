import type { Device } from '@/gen/v1/device_pb'
import { Badge } from '@/components/ui/badge'
import type { PingDataPoint } from '../../types'
import { DevicePingSparkline } from './device-ping-sparkline'
import { useDevicesContext } from '../devices-provider'
import { Activity } from 'lucide-react'

interface DeviceCardPingProps {
  device: Device
  pingMs: number
  pingHistory: PingDataPoint[]
  pingTarget: string
}

export function DeviceCardPing({
  device,
  pingMs,
  pingHistory,
  pingTarget,
}: DeviceCardPingProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()

  const openAnalytics = () => {
    setCurrentRow(device)
    setOpen('ping-analytics')
  }

  const openSettings = () => {
    setCurrentRow(device)
    setOpen('ping-settings')
  }

  return (
    <section className='py-2.5 border-b flex items-center justify-between gap-2'>
      <div className='flex items-center gap-1.5 text-xs min-w-0 flex-1'>
        {/* Latency Readout Button */}
        <button
          type='button'
          onClick={openAnalytics}
          className='flex items-baseline gap-1 hover:opacity-80 transition-opacity text-left cursor-pointer'
          title='Buka Analisis Ping Historis'
        >
          <Activity className='h-3.5 w-3.5 text-muted-foreground shrink-0 self-center' />
          <span className='font-mono font-bold text-sm text-foreground'>
            {pingMs}
          </span>
          <span className='text-[10px] text-muted-foreground'>ms</span>
          <span className='text-[10px] text-muted-foreground ml-1 truncate max-w-[110px] sm:max-w-[140px]'>
            → <span className='font-mono text-primary font-medium'>{pingTarget}</span>
          </span>
        </button>

        {/* TimescaleDB Telemetry Recording Status */}
        {device.pingEnabled ? (
          <Badge
            variant='outline'
            className='text-[9px] px-1.5 py-0 bg-emerald-500/10 text-emerald-600 border-emerald-300 dark:border-emerald-700/50 cursor-pointer hover:bg-emerald-500/20 transition-colors flex items-center gap-1 shrink-0'
            onClick={openSettings}
            title='Metrik Aktif Direkam ke Database (Klik untuk Ubah)'
          >
            <span className='h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse' />
            REC
          </Badge>
        ) : (
          <Badge
            variant='outline'
            className='text-[9px] px-1.5 py-0 bg-muted/50 text-muted-foreground cursor-pointer hover:bg-accent transition-colors shrink-0'
            onClick={openSettings}
            title='Perekaman Metrik Tidak Aktif (Klik untuk Aktifkan)'
          >
            OFF
          </Badge>
        )}
      </div>

      {/* Sparkline Canvas Container */}
      <div
        className='w-28 sm:w-32 h-7 bg-muted/30 rounded overflow-hidden cursor-pointer hover:opacity-90 border border-border/40 shrink-0 transition-opacity'
        onClick={openAnalytics}
        title='Klik untuk Melihat Grafik Ping Lengkap'
      >
        <DevicePingSparkline pingHistory={pingHistory} />
      </div>
    </section>
  )
}

