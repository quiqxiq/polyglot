import { Link } from '@tanstack/react-router'
import { Plus, Upload } from 'lucide-react'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { Button } from '@/components/ui/button'
import { usePlans } from './plans-provider'

export function PlansPrimaryButtons() {
  const { setOpen } = usePlans()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'billing:manage')

  if (!canManage) return null

  return (
    <div className='flex items-center gap-2'>
      <Button variant='outline' className='space-x-1' asChild>
        <Link to='/plans/import'>
          <span>Import Paket</span> <Upload size={18} />
        </Link>
      </Button>
      <Button className='space-x-1' onClick={() => setOpen('create')}>
        <span>Tambah Paket</span> <Plus size={18} />
      </Button>
    </div>
  )
}
