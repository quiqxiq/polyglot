import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { usePlans } from './plans-provider'

export function PlansPrimaryButtons() {
  const { setOpen } = usePlans()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'billing:manage')

  if (!canManage) return null

  return (
    <Button className='space-x-1' onClick={() => setOpen('create')}>
      <span>Tambah Paket</span> <Plus size={18} />
    </Button>
  )
}
