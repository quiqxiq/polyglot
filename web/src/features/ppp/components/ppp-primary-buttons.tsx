import { Button } from '@/components/ui/button'
import { ShieldPlus, UserPlus } from 'lucide-react'
import { usePPP } from '../context/ppp-context'

export function PPPPrimaryButtons() {
  const { setOpen, setCurrentSecret, setCurrentProfile } = usePPP()

  const handleAddSecret = () => {
    setCurrentSecret(null)
    setOpen('secret-create')
  }

  const handleAddProfile = () => {
    setCurrentProfile(null)
    setOpen('profile-create')
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={handleAddProfile}
        className="h-8 sm:h-9 text-xs sm:text-sm"
      >
        <ShieldPlus className="mr-1.5 sm:mr-2 h-3.5 w-3.5 sm:h-4 sm:w-4" />
        Add Profile
      </Button>

      <Button
        size="sm"
        onClick={handleAddSecret}
        className="h-8 sm:h-9 text-xs sm:text-sm"
      >
        <UserPlus className="mr-1.5 sm:mr-2 h-3.5 w-3.5 sm:h-4 sm:w-4" />
        Add Secret
      </Button>
    </div>
  )
}
