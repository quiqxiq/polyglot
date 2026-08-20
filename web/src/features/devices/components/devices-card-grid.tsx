import { Device } from '@/gen/v1/device_pb'
import { DeviceCard } from './device-card'
import { PlusIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { useDevicesContext } from './devices-provider'

interface DevicesCardGridProps {
  devices: Device[]
  isLoading?: boolean
  searchTerm?: string
}

export function DevicesCardGrid({ devices, isLoading, searchTerm = '' }: DevicesCardGridProps) {
  const { setOpen } = useDevicesContext()

  if (isLoading) {
    return (
      <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3 pt-2 pb-16'>
        {[1, 2, 3].map((i) => (
          <div key={i} className='h-96 rounded-xl border bg-card/50 p-4 animate-pulse' />
        ))}
      </div>
    )
  }

  if (devices.length === 0) {
    if (searchTerm) {
      return (
        <div className='flex flex-col items-center justify-center py-16 text-center border rounded-lg bg-muted/20 my-4'>
          <p className='text-base font-medium text-foreground'>No matching devices found</p>
          <p className='text-sm text-muted-foreground mt-1'>
            No devices matched your search query &quot;{searchTerm}&quot;.
          </p>
        </div>
      )
    }

    return (
      <div className='flex flex-col items-center justify-center py-16 text-center border rounded-lg bg-muted/20 my-4 space-y-3'>
        <p className='text-base font-medium text-foreground'>No devices in inventory</p>
        <p className='text-sm text-muted-foreground max-w-sm'>
          Add your first MikroTik router or network device to begin real-time monitoring and management.
        </p>
        <Button onClick={() => setOpen('add')} className='gap-2'>
          <PlusIcon className='h-4 w-4' />
          Add First Device
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
