import { UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRegistration } from './registration-provider'

export function RegistrationPrimaryButtons() {
  const { setOpen, setCurrentRow } = useRegistration()

  const handleCreate = () => {
    setCurrentRow(null)
    setOpen('submit')
  }

  return (
    <div className='flex items-center gap-2'>
      <Button onClick={handleCreate} className='gap-1.5'>
        <UserPlus className='h-4 w-4' />
        Daftar Pelanggan Baru
      </Button>
    </div>
  )
}
