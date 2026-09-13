import {
  Copy,
  MoreHorizontal,
  Network,
  Pencil,
  Router,
  Server,
  Wifi,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { toast } from 'sonner'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Plan } from '@/gen/v1/plan_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import {
  PROVISION_STATUS_META,
  subscriptionStatusBadge,
} from '@/features/billing/data/constants'
import { formatBandwidthKbps, formatCurrency, formatUnixDate } from '@/lib/format'

export type SubscriptionLifecycleAction =
  | 'resume'
  | 'terminate'
  | 'activate'
  | 'isolate'
  | 'restore'

interface SubscriptionCardProps {
  subscription: Subscription
  plan?: Plan
  deviceName: string
  onEdit: (subscription: Subscription) => void
  onChangePlan: (subscription: Subscription) => void
  onSuspend: (subscription: Subscription) => void
  onLifecycle: (action: SubscriptionLifecycleAction, subscription: Subscription) => void
}

function serviceTypeMeta(serviceType: string) {
  if (serviceType === 'HOTSPOT') {
    return {
      label: 'Hotspot',
      icon: Wifi,
      className: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
    }
  }
  if (serviceType === 'DEDICATED') {
    return {
      label: 'Dedicated',
      icon: Server,
      className: 'bg-orange-500/10 text-orange-600 dark:text-orange-400',
    }
  }
  return {
    label: 'PPPoE',
    icon: Network,
    className: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  }
}

function DetailField({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <p className='text-[10px] uppercase tracking-wide text-muted-foreground'>{label}</p>
      <p className={`mt-0.5 truncate text-sm font-semibold ${mono ? 'font-mono' : ''}`}>
        {value}
      </p>
    </div>
  )
}

export function SubscriptionCard({
  subscription,
  plan,
  deviceName,
  onEdit,
  onChangePlan,
  onSuspend,
  onLifecycle,
}: SubscriptionCardProps) {
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canBilling = canPermission(permissions, 'billing:manage')

  const typeMeta = serviceTypeMeta(subscription.serviceType)
  const statusBadge = subscriptionStatusBadge(subscription.status)
  const provisionMeta =
    PROVISION_STATUS_META[subscription.provisionStatus as keyof typeof PROVISION_STATUS_META] ||
    PROVISION_STATUS_META.NONE

  const isTerminated = subscription.status === 'TERMINATED'
  const rateLimit =
    subscription.pppoeConfig?.rateLimit ||
    subscription.hotspotConfig?.rateLimit ||
    subscription.rateLimit ||
    formatBandwidthKbps(plan?.bandwidthDownloadKbps, plan?.bandwidthUploadKbps) ||
    '-'

  const effectivePrice =
    subscription.customPrice > 0
      ? subscription.customPrice
      : subscription.planPrice > 0
        ? subscription.planPrice
        : (plan?.price ?? 0)

  const staticIp =
    subscription.pppoeConfig?.remoteAddress || subscription.hotspotConfig?.ipAddress

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} berhasil disalin`)
  }

  return (
    <Card className='overflow-hidden shadow-xs transition-colors hover:border-primary/40'>
      <CardHeader className='gap-0 border-b bg-muted/15 p-4'>
        <div className='flex items-start justify-between gap-3'>
          <div className='flex min-w-0 items-center gap-3'>
            <div
              className={`flex size-10 shrink-0 items-center justify-center rounded-xl ${typeMeta.className}`}
            >
              <typeMeta.icon className='h-5 w-5' />
            </div>
            <div className='min-w-0'>
              <div className='flex items-center gap-1.5'>
                <p className='truncate text-sm font-bold'>
                  {subscription.planName || plan?.name || 'Paket Internet'}
                </p>
                <Badge variant='outline' className='text-[10px]'>
                  {typeMeta.label}
                </Badge>
              </div>
              <button
                type='button'
                onClick={() =>
                  copyToClipboard(subscription.remoteUsername, 'Username')
                }
                className='mt-0.5 inline-flex items-center gap-1 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground'
                title='Salin username'
              >
                {subscription.remoteUsername}
                <Copy className='h-3 w-3' />
              </button>
            </div>
          </div>

          <div className='flex shrink-0 flex-col items-end gap-1.5'>
            <Badge variant='outline' className={`text-[10px] ${statusBadge.className}`}>
              {statusBadge.label}
            </Badge>
            <Badge variant='outline' className={`text-[10px] ${provisionMeta.className}`}>
              {provisionMeta.label}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className='grid grid-cols-2 gap-3 p-4 sm:grid-cols-4'>
        <DetailField label='Router' value={deviceName || '-'} />
        <DetailField label='Kecepatan' value={rateLimit} mono />
        <DetailField
          label='Biaya Bulanan'
          value={`${formatCurrency(effectivePrice)}${subscription.customPrice > 0 ? ' *' : ''}`}
        />
        <DetailField
          label='Jatuh Tempo'
          value={`Tgl ${subscription.billingDay || 1} · ${subscription.billingCycle || 'MONTHLY'}`}
        />

        {(staticIp ||
          subscription.pppoeConfig?.callerId ||
          subscription.hotspotConfig?.macAddress ||
          (subscription.startDateUnix > 0 && !isTerminated)) && (
          <div className='col-span-2 flex flex-wrap items-center gap-x-4 gap-y-1.5 border-t pt-3 text-xs text-muted-foreground sm:col-span-4'>
            {staticIp && (
              <span className='inline-flex items-center gap-1.5 font-mono'>
                IP Statis:
                <span className='font-semibold text-foreground'>{staticIp}</span>
                <Button
                  size='icon'
                  variant='ghost'
                  className='h-5 w-5 text-muted-foreground hover:text-foreground'
                  onClick={() => copyToClipboard(staticIp, 'IP Address')}
                  title='Salin IP'
                >
                  <Copy className='h-2.5 w-2.5' />
                </Button>
              </span>
            )}
            {subscription.pppoeConfig?.callerId && (
              <span className='font-mono'>
                Caller ID: {subscription.pppoeConfig.callerId}
              </span>
            )}
            {subscription.hotspotConfig?.macAddress && (
              <span className='font-mono'>
                MAC: {subscription.hotspotConfig.macAddress}
              </span>
            )}
            {subscription.startDateUnix > 0 && !isTerminated && (
              <span>Mulai: {formatUnixDate(subscription.startDateUnix)}</span>
            )}
          </div>
        )}
      </CardContent>

      <CardFooter className='flex items-center justify-between gap-2 border-t bg-muted/20 px-4 py-2.5'>
        <span className='flex items-center gap-1 text-[11px] text-muted-foreground'>
          <Router className='h-3 w-3' />
          {deviceName || 'Belum ditugaskan'}
        </span>

        {canBilling && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant='outline' size='sm' className='h-7 gap-1 text-xs'>
                Kelola
                <MoreHorizontal className='h-3.5 w-3.5' />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end' className='w-52'>
              <DropdownMenuItem onSelect={() => onEdit(subscription)}>
                <Pencil className='mr-2 h-4 w-4' />
                Edit Layanan
              </DropdownMenuItem>
              {!isTerminated && (
                <DropdownMenuItem onSelect={() => onChangePlan(subscription)}>
                  <Network className='mr-2 h-4 w-4' />
                  Ubah Paket
                </DropdownMenuItem>
              )}

              {subscription.status === 'ACTIVE' && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onSelect={() => onLifecycle('isolate', subscription)}>
                    Isolir Pelanggan
                  </DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => onSuspend(subscription)}>
                    Tangguhkan
                  </DropdownMenuItem>
                </>
              )}

              {subscription.status === 'ISOLATED' && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onSelect={() => onLifecycle('restore', subscription)}>
                    Buka Isolir
                  </DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => onSuspend(subscription)}>
                    Tangguhkan
                  </DropdownMenuItem>
                </>
              )}

              {subscription.status === 'SUSPENDED' && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onSelect={() => onLifecycle('resume', subscription)}>
                    Pulihkan Layanan
                  </DropdownMenuItem>
                </>
              )}

              {subscription.provisionStatus !== 'OK' && !isTerminated && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onSelect={() => onLifecycle('activate', subscription)}>
                    Aktifkan Provisi
                  </DropdownMenuItem>
                </>
              )}

              {!isTerminated && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onSelect={() => onLifecycle('terminate', subscription)}
                    className='text-destructive focus:text-destructive'
                  >
                    Terminate Langganan
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </CardFooter>
    </Card>
  )
}
