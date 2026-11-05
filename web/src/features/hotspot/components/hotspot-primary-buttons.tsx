import { Plus, Ticket, Clock, UserPlus, Search, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useCan } from '@/hooks/use-can'
import { useHotspot } from '../context/hotspot-context'
import { useDeviceStore } from '@/stores/device-store'

export function HotspotPrimaryButtons() {
  const { setOpen, setCurrentUser, setCurrentProfile } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()

  const canCheckVoucher = useCan('hotspot:voucher:generate')
  const canBulkClean = useCan('hotspot:user:write')
  const canExpireMonitor = useCan('hotspot:expire:read') || useCan('hotspot:expire:manage')
  const canGenerateVouchers = useCan('hotspot:voucher:generate')
  const canAddProfile = useCan('hotspot:profile:write')
  const canAddUser = useCan('hotspot:user:write')

  const hasAnyAction =
    canCheckVoucher ||
    canBulkClean ||
    canExpireMonitor ||
    canGenerateVouchers ||
    canAddProfile ||
    canAddUser

  if (!hasAnyAction) {
    return null
  }

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

  const handleQuickChecker = () => {
    setOpen('quick-voucher-checker')
  }

  const handleBulkCleaner = () => {
    setOpen('user-bulk-cleaner')
  }

  return (
    <div className='flex flex-wrap items-center gap-2'>
      {canCheckVoucher && (
        <Button
          variant='outline'
          size='sm'
          onClick={handleQuickChecker}
          disabled={!selectedDeviceId}
          className='gap-1.5 h-9'
          title='Inspect voucher validity and live status'
        >
          <Search className='size-4 text-sky-500' />
          Check Voucher
        </Button>
      )}

      {canBulkClean && (
        <Button
          variant='outline'
          size='sm'
          onClick={handleBulkCleaner}
          disabled={!selectedDeviceId}
          className='gap-1.5 h-9 text-destructive border-destructive/30 hover:bg-destructive/10'
          title='Bulk purge users by profile, batch tag, or expiration'
        >
          <Trash2 className='size-4' />
          Bulk Cleaner
        </Button>
      )}

      {canExpireMonitor && (
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
      )}

      {canGenerateVouchers && (
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
      )}

      {canAddProfile && (
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
      )}

      {canAddUser && (
        <Button
          size='sm'
          onClick={handleAddUser}
          disabled={!selectedDeviceId}
          className='gap-1.5 h-9'
        >
          <UserPlus className='size-4' />
          Add User
        </Button>
      )}
    </div>
  )
}
