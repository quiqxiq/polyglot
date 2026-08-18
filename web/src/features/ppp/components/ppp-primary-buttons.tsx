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
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={handleAddProfile}
        className="h-9"
      >
        <ShieldPlus className="mr-2 h-4 w-4" />
        Add Profile
      </Button>

      <Button size="sm" onClick={handleAddSecret} className="h-9">
        <UserPlus className="mr-2 h-4 w-4" />
        Add Secret
      </Button>
    </div>
  )
}
