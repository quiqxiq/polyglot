import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useSubscriptions } from './subscriptions-provider'

export function SubscriptionsPrimaryButtons() {
  const { setOpen } = useSubscriptions()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'billing:manage')

  if (!canManage) return null

  return (
    <Button onClick={() => setOpen('create')} size='sm'>
      <Plus className='mr-2 h-4 w-4' />
      Tambah Langganan
    </Button>
  )
}
