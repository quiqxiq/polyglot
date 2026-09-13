import { Plus, Repeat } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Plan } from '@/gen/v1/plan_pb'
import type { Device } from '@/gen/v1/device_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import {
  SubscriptionCard,
  type SubscriptionLifecycleAction,
} from './subscription-card'

interface CustomerSubscriptionsTabProps {
  subscriptions: Subscription[]
  plans: Plan[]
  devices: Device[]
  isLoading: boolean
  onAddSubscription: () => void
  onEdit: (subscription: Subscription) => void
  onChangePlan: (subscription: Subscription) => void
  onSuspend: (subscription: Subscription) => void
  onLifecycle: (
    action: SubscriptionLifecycleAction,
    subscription: Subscription
  ) => void
}

export function CustomerSubscriptionsTab({
  subscriptions,
  plans,
  devices,
  isLoading,
  onAddSubscription,
  onEdit,
  onChangePlan,
  onSuspend,
  onLifecycle,
}: CustomerSubscriptionsTabProps) {
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canBilling = canPermission(permissions, 'billing:manage')

  const planById = new Map(plans.map((plan) => [plan.id, plan]))
  const deviceNameById = new Map(
    devices.filter((device) => device.id && device.name).map((device) => [device.id, device.name])
  )

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div>
          <h2 className='text-sm font-semibold'>Layanan Internet</h2>
          <p className='text-xs text-muted-foreground'>
            Kelola akun PPPoE, Hotspot, dan Dedicated beserta status provisinya.
          </p>
        </div>
        {canBilling && (
          <Button size='sm' className='h-8 gap-1.5 text-xs' onClick={onAddSubscription}>
            <Plus className='h-3.5 w-3.5' />
            Tambah Layanan
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className='space-y-3'>
          {Array.from({ length: 2 }).map((_, index) => (
            <Skeleton key={index} className='h-44 w-full rounded-xl' />
          ))}
        </div>
      ) : subscriptions.length === 0 ? (
        <div className='rounded-xl border border-dashed py-14 text-center'>
          <Repeat className='mx-auto h-10 w-10 text-muted-foreground/40' />
          <p className='mt-3 text-sm font-semibold'>Belum Ada Layanan</p>
          <p className='mx-auto mt-1 max-w-xs text-xs text-muted-foreground'>
            Pelanggan ini belum memiliki akun PPPoE, Hotspot, atau Dedicated.
          </p>
          {canBilling && (
            <Button
              variant='outline'
              size='sm'
              className='mt-4 gap-1.5'
              onClick={onAddSubscription}
            >
              <Plus className='h-3.5 w-3.5' />
              Buat Langganan Pertama
            </Button>
          )}
        </div>
      ) : (
        <div className='space-y-3'>
          {subscriptions.map((subscription) => (
            <SubscriptionCard
              key={subscription.id}
              subscription={subscription}
              plan={planById.get(subscription.planId)}
              deviceName={
                subscription.deviceName ||
                deviceNameById.get(subscription.deviceId) ||
                subscription.deviceId
              }
              onEdit={onEdit}
              onChangePlan={onChangePlan}
              onSuspend={onSuspend}
              onLifecycle={onLifecycle}
            />
          ))}
        </div>
      )}
    </div>
  )
}
