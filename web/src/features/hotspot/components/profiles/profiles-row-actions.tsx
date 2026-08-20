import { MoreHorizontal, Pencil, Trash2, Ticket } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { HotspotProfile } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

interface ProfilesRowActionsProps {
  profile: HotspotProfile
}

export function ProfilesRowActions({ profile }: ProfilesRowActionsProps) {
  const { setOpen, setCurrentProfile } = useHotspot()

  const handleEdit = () => {
    setCurrentProfile(profile)
    setOpen('profile-update')
  }

  const handleDelete = () => {
    setCurrentProfile(profile)
    setOpen('profile-delete')
  }

  const handleGenerate = () => {
    setCurrentProfile(profile)
    setOpen('voucher-generate')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0'>
          <span className='sr-only'>Open menu</span>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-48'>
        <DropdownMenuItem onClick={handleGenerate}>
          <Ticket className='mr-2 h-4 w-4 text-primary' />
          Generate Vouchers
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleEdit}>
          <Pencil className='mr-2 h-4 w-4' />
          Edit Profile
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleDelete} className='text-destructive focus:text-destructive'>
          <Trash2 className='mr-2 h-4 w-4' />
          Delete Profile
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
