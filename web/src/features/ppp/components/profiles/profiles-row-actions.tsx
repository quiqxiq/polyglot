import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { PPPProfile } from '@/gen/v1/ppp_pb'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { Edit2, Trash2 } from 'lucide-react'
import { usePPP } from '../../context/ppp-context'

interface ProfilesRowActionsProps {
  row: PPPProfile
}

export function ProfilesRowActions({ row }: ProfilesRowActionsProps) {
  const { setOpen, setCurrentProfile } = usePPP()

  const handleEdit = () => {
    setCurrentProfile(row)
    setOpen('profile-update')
  }

  const handleDelete = () => {
    setCurrentProfile(row)
    setOpen('profile-delete')
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
        >
          <DotsHorizontalIcon className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[160px]">
        <DropdownMenuItem onClick={handleEdit}>
          <Edit2 className="mr-2 h-4 w-4" />
          Edit Profile
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={handleDelete}
          className="text-destructive focus:text-destructive"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete Profile
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
