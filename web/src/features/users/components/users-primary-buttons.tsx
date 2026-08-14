import { UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useUsers } from './users-provider'

export function UsersPrimaryButtons() {
  const { setOpen } = useUsers()
  return (
    <Button className='space-x-1' onClick={() => setOpen('create')}>
      <span>Add User</span> <UserPlus size={18} />
    </Button>
  )
}
