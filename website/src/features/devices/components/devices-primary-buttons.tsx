import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useDevicesContext } from './devices-provider'

export function DevicesPrimaryButtons() {
  const { setOpen, setCurrentRow } = useDevicesContext()

  return (
    <div className='flex items-center gap-2'>
      <Button
        onClick={() => {
          setCurrentRow(null)
          setOpen('add')
        }}
        className='gap-1.5'
      >
        <Plus className='h-4 w-4' />
        Add Device
      </Button>
    </div>
  )
}
