import { Link } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { Monitor, Network, Wifi } from 'lucide-react'

export function DeviceFleetCard() {
  const { data: devices = [], isLoading } = useDevicesQuery()
  const { selectedDeviceId, setSelectedDeviceId } = useDeviceStore()

  return (
    <Card className='col-span-4 shadow-xs'>
      <CardHeader className='flex flex-row items-center justify-between pb-3'>
        <div>
          <CardTitle className='text-base font-semibold'>Armada Router MikroTik (Fleet)</CardTitle>
          <CardDescription>
            {devices.length > 0
              ? `${devices.length} router terhubung dalam NetOps Engine`
              : 'Daftar router yang terdaftar untuk manajemen jaringan'}
          </CardDescription>
        </div>
        <Button asChild size='sm' variant='outline' className='h-8 text-xs'>
          <Link to='/devices'>Kelola Router</Link>
        </Button>
      </CardHeader>
      <CardContent className='pt-1'>
        {isLoading ? (
          <div className='space-y-2.5'>
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className='h-14 w-full rounded-md' />
            ))}
          </div>
        ) : devices.length === 0 ? (
          <div className='flex h-36 flex-col items-center justify-center text-center text-xs text-muted-foreground'>
            <Monitor className='size-8 text-muted-foreground/40 mb-1.5' />
            <p>Belum ada router yang ditambahkan.</p>
            <Button asChild size='sm' variant='outline' className='mt-2.5 h-7 text-xs'>
              <Link to='/devices'>+ Tambah Router Pertama</Link>
            </Button>
          </div>
        ) : (
          <div className='divide-y divide-border rounded-md border text-xs'>
            {devices.map((device) => {
              const isSelected = selectedDeviceId === device.id
              return (
                <div
                  key={device.id}
                  className={`flex flex-wrap items-center justify-between gap-2 p-3 transition-colors hover:bg-muted/50 ${
                    isSelected ? 'bg-primary/5' : ''
                  }`}
                >
                  <div className='flex items-center gap-3 min-w-0'>
                    <div
                      className={`flex size-8 shrink-0 items-center justify-center rounded-md ${
                        device.enabled
                          ? 'bg-emerald-500/15 text-emerald-600'
                          : 'bg-muted text-muted-foreground'
                      }`}
                    >
                      <Monitor className='size-4' />
                    </div>
                    <div className='min-w-0'>
                      <div className='flex items-center gap-1.5'>
                        <span className='font-semibold text-foreground truncate'>
                          {device.name}
                        </span>
                        {isSelected && (
                          <Badge variant='default' className='h-4 px-1 text-[9px] font-medium'>
                            Aktif
                          </Badge>
                        )}
                        <span
                          className={`size-2 rounded-full ${
                            device.enabled ? 'bg-emerald-500' : 'bg-muted-foreground'
                          }`}
                          title={device.enabled ? 'Enabled' : 'Disabled'}
                        />
                      </div>
                      <p className='text-muted-foreground text-[11px] font-mono'>
                        {device.host}:{device.port} • {device.vendor || 'mikrotik'}
                      </p>
                    </div>
                  </div>

                  <div className='flex items-center gap-1.5 ms-auto'>
                    {!isSelected && (
                      <Button
                        size='sm'
                        variant='ghost'
                        className='h-7 px-2 text-[11px]'
                        onClick={() => setSelectedDeviceId(device.id)}
                      >
                        Pilih Router
                      </Button>
                    )}
                    <Button asChild size='sm' variant='outline' className='h-7 px-2 text-[11px] gap-1'>
                      <Link
                        to='/hotspot'
                        search={{ tab: 'users' }}
                        onClick={() => setSelectedDeviceId(device.id)}
                      >
                        <Wifi className='size-3' /> Hotspot
                      </Link>
                    </Button>
                    <Button asChild size='sm' variant='outline' className='h-7 px-2 text-[11px] gap-1'>
                      <Link
                        to='/ppp'
                        search={{ tab: 'secrets' }}
                        onClick={() => setSelectedDeviceId(device.id)}
                      >
                        <Network className='size-3' /> PPPoE
                      </Link>
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
