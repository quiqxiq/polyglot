import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { HotspotHost } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

interface HostsRowActionsProps {
  host: HotspotHost
}

export function HostsRowActions({ host }: HostsRowActionsProps) {
  const { setOpen, setCurrentHost } = useHotspot()

  const handleDelete = () => {
    setCurrentHost(host)
    setOpen('host-delete')
  }

  return (
    <Button
      variant='ghost'
      size='sm'
      onClick={handleDelete}
      className='h-8 px-2 text-destructive hover:text-destructive hover:bg-destructive/10'
      title='Remove host'
    >
      <Trash2 className='size-3.5 mr-1' />
      Remove
    </Button>
  )
}
