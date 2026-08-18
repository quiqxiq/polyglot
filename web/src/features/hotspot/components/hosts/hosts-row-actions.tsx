import { Trash2, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { HotspotHost } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

interface HostsRowActionsProps {
  host: HotspotHost
}

export function HostsRowActions({ host }: HostsRowActionsProps) {
  const { setOpen, setCurrentHost, setPrefillBinding } = useHotspot()

  const handleDelete = () => {
    setCurrentHost(host)
    setOpen('host-delete')
  }

  const handleMakeBinding = () => {
    setPrefillBinding({
      macAddress: host.macAddress,
      address: host.address,
      server: host.server,
    })
    setOpen('binding-create')
  }

  return (
    <div className='flex items-center gap-1'>
      <Button
        variant='outline'
        size='sm'
        onClick={handleMakeBinding}
        className='h-7 px-2 text-xs gap-1 text-emerald-600 dark:text-emerald-400 hover:text-emerald-700'
        title='Make IP Binding (Bypass)'
      >
        <ShieldCheck className='size-3.5' />
        Bypass
      </Button>

      <Button
        variant='ghost'
        size='sm'
        onClick={handleDelete}
        className='h-7 px-2 text-xs text-destructive hover:text-destructive hover:bg-destructive/10'
        title='Remove host'
      >
        <Trash2 className='size-3.5 mr-1' />
        Remove
      </Button>
    </div>
  )
}
