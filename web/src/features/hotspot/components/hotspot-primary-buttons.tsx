import { Plus, Ticket, Clock, UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useHotspot } from '../context/hotspot-context'
import { useDeviceStore } from '@/stores/device-store'

export function HotspotPrimaryButtons() {
  const { setOpen, setCurrentUser, setCurrentProfile } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()

  const handleAddUser = () => {
    setCurrentUser(null)
    setOpen('user-create')
  }

  const handleAddProfile = () => {
    setCurrentProfile(null)
    setOpen('profile-create')
  }

  const handleGenerate = () => {
    setCurrentProfile(null)
    setOpen('voucher-generate')
  }

  const handleExpireMonitor = () => {
    setOpen('expire-monitor')
  }

  return (
    <div className='flex flex-wrap items-center gap-2'>
      <Button
        variant='outline'
        size='sm'
        onClick={handleExpireMonitor}
        disabled={!selectedDeviceId}
        className='gap-1.5 h-9'
      >
        <Clock className='size-4 text-amber-500' />
        Expire Monitor
      </Button>

      <Button
        variant='outline'
        size='sm'
        onClick={handleGenerate}
        disabled={!selectedDeviceId}
        className='gap-1.5 h-9 text-primary border-primary hover:bg-primary/10'
      >
        <Ticket className='size-4' />
        Generate Vouchers
      </Button>

      <Button
        variant='outline'
        size='sm'
        onClick={handleAddProfile}
        disabled={!selectedDeviceId}
        className='gap-1.5 h-9'
      >
        <Plus className='size-4' />
        Add Profile
      </Button>

      <Button
        size='sm'
        onClick={handleAddUser}
        disabled={!selectedDeviceId}
        className='gap-1.5 h-9'
      >
        <UserPlus className='size-4' />
        Add User
      </Button>
    </div>
  )
}
