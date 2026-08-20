import { MoreHorizontal, Pencil, RotateCcw, Trash2, Printer } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { HotspotUser } from '@/gen/v1/hotspot_pb'
import { useHotspot } from '../../context/hotspot-context'

interface UsersRowActionsProps {
  user: HotspotUser
}

export function UsersRowActions({ user }: UsersRowActionsProps) {
  const { setOpen, setCurrentUser, setPrintSingleUserId, setPrintBatchComment } = useHotspot()

  const handleEdit = () => {
    setCurrentUser(user)
    setOpen('user-update')
  }

  const handleReset = () => {
    setCurrentUser(user)
    setOpen('user-reset')
  }

  const handleDelete = () => {
    setCurrentUser(user)
    setOpen('user-delete')
  }

  const handlePrint = () => {
    setPrintSingleUserId(user.id)
    setPrintBatchComment('')
    setOpen('voucher-print')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0'>
          <span className='sr-only'>Open menu</span>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-44'>
        <DropdownMenuItem onClick={handlePrint}>
          <Printer className='mr-2 h-4 w-4 text-primary' />
          Print Voucher
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleEdit}>
          <Pencil className='mr-2 h-4 w-4' />
          Edit User
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleReset}>
          <RotateCcw className='mr-2 h-4 w-4 text-amber-500' />
          Reset Counters
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleDelete} className='text-destructive focus:text-destructive'>
          <Trash2 className='mr-2 h-4 w-4' />
          Delete User
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
