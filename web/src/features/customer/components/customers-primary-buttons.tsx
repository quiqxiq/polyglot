import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useCustomers } from './customers-provider'

export function CustomersPrimaryButtons() {
  const { setOpen } = useCustomers()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'customer:manage')

  if (!canManage) return null

  return (
    <Button className='space-x-1' onClick={() => setOpen('create')}>
      <span>Tambah Pelanggan</span> <Plus size={18} />
    </Button>
  )
}
