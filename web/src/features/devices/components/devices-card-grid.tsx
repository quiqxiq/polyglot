import type { Device } from '@/gen/v1/device_pb'
import { DeviceCard } from './card/device-card'
import { PlusIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useDevicesContext } from './devices-provider'
import { Server } from 'lucide-react'

interface DevicesCardGridProps {
  devices: Device[]
  isLoading?: boolean
  searchTerm?: string
}

export function DevicesCardGrid({
  devices,
  isLoading,
  searchTerm = '',
}: DevicesCardGridProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()

  if (isLoading) {
    return (
      <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3 pt-2 pb-16'>
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div
            key={i}
            className='flex flex-col rounded-xl border bg-card/60 p-4 space-y-4 shadow-2xs'
          >
            {/* Header Skeleton */}
            <div className='flex items-start justify-between gap-2 border-b pb-3'>
              <div className='flex items-center gap-2.5 flex-1'>
                <Skeleton className='h-3 w-3 rounded-full' />
                <div className='space-y-1.5 flex-1'>
                  <Skeleton className='h-4 w-32' />
                  <Skeleton className='h-3 w-24' />
                </div>
              </div>
              <Skeleton className='h-7 w-28 rounded-md' />
            </div>

            {/* Metrics Skeleton */}
            <div className='grid grid-cols-2 gap-3 py-2 border-b'>
              <div className='space-y-1.5'>
                <Skeleton className='h-3 w-16' />
                <Skeleton className='h-2 w-full rounded-full' />
              </div>
              <div className='space-y-1.5'>
                <Skeleton className='h-3 w-16' />
                <Skeleton className='h-2 w-full rounded-full' />
              </div>
            </div>

            {/* Ping Sparkline Skeleton */}
            <div className='flex items-center justify-between gap-2 py-2 border-b'>
              <Skeleton className='h-4 w-28' />
              <Skeleton className='h-7 w-28 rounded' />
            </div>

            {/* Traffic Chart Skeleton */}
            <div className='space-y-2 pt-1'>
              <div className='flex items-center justify-between'>
                <Skeleton className='h-6 w-24' />
                <Skeleton className='h-4 w-20' />
              </div>
              <Skeleton className='h-24 w-full rounded-lg' />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (devices.length === 0) {
    if (searchTerm) {
      return (
        <div className='flex flex-col items-center justify-center py-16 text-center border rounded-xl bg-muted/20 my-4 p-6 shadow-2xs'>
          <Server className='h-10 w-10 text-muted-foreground/50 mb-2' />
          <p className='text-base font-semibold text-foreground'>
            Perangkat tidak ditemukan
          </p>
          <p className='text-sm text-muted-foreground mt-1 max-w-md'>
            Tidak ada perangkat yang sesuai dengan kata kunci pencarian &quot;
            <span className='font-mono text-foreground'>{searchTerm}</span>&quot;.
          </p>
        </div>
      )
    }

    return (
      <div className='flex flex-col items-center justify-center py-16 text-center border rounded-xl bg-muted/20 my-4 space-y-3 p-6 shadow-2xs'>
        <div className='rounded-full bg-primary/10 p-3 text-primary'>
          <Server className='h-8 w-8' />
        </div>
        <p className='text-lg font-semibold text-foreground'>
          Belum ada perangkat dalam inventaris
        </p>
        <p className='text-sm text-muted-foreground max-w-sm'>
          Tambahkan router MikroTik, Cisco, atau perangkat jaringan pertama Anda untuk
          memulai monitoring dan manajemen realtime.
        </p>
        <Button
          onClick={() => {
            setCurrentRow(null)
            setOpen('add')
          }}
          className='gap-2 mt-2'
        >
          <PlusIcon className='h-4 w-4' />
          Tambah Perangkat Pertama
        </Button>
      </div>
    )
  }

  return (
    <div className='grid gap-4 pt-2 pb-16 md:grid-cols-2 lg:grid-cols-3'>
      {devices.map((device) => (
        <DeviceCard key={device.id} device={device} />
      ))}
    </div>
  )
}
