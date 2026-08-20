import { LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { HotspotActiveSession } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

interface ActiveRowActionsProps {
  session: HotspotActiveSession
}

export function ActiveRowActions({ session }: ActiveRowActionsProps) {
  const { setOpen, setCurrentSession } = useHotspot()

  const handleKick = () => {
    setCurrentSession(session)
    setOpen('session-kick')
  }

  return (
    <Button
      variant='ghost'
      size='sm'
      onClick={handleKick}
      className='h-8 px-2 text-destructive hover:text-destructive hover:bg-destructive/10'
      title='Kick / Disconnect session'
    >
      <LogOut className='size-3.5 mr-1' />
      Kick
    </Button>
  )
}
