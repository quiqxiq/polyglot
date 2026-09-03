import {
  Wifi,
  Network,
  Server,
  Plus,
  Copy,
  Pencil,
  Repeat,
  Gauge,
  Router,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Plan } from '@/gen/v1/plan_pb'
import type { Device } from '@/gen/v1/device_pb'
import { subscriptionStatusBadge, PROVISION_STATUS_META } from '@/features/billing/data/constants'

interface CustomerSubscriptionsTabProps {
  subscriptions: Subscription[]
  plans: Plan[]
  devices: Device[]
  isLoading: boolean
  onAddSubscription: () => void
  onEditSubscription: (sub: Subscription) => void
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatPlanRate(downloadKbps?: number, uploadKbps?: number) {
  if (!downloadKbps && !uploadKbps) return ''
  const formatSide = (kbps?: number) => {
    if (!kbps || kbps <= 0) return ''
    if (kbps >= 1000) return `${Math.round(kbps / 1000)}M`
    return `${kbps}k`
  }
  const dl = formatSide(downloadKbps)
  const ul = formatSide(uploadKbps)
  if (!dl && !ul) return ''
  return `${dl || '0'}/${ul || '0'}`
}

export function CustomerSubscriptionsTab({
  subscriptions,
  plans,
  devices,
  isLoading,
  onAddSubscription,
  onEditSubscription,
}: CustomerSubscriptionsTabProps) {
  const planById = new Map(plans.map((p) => [p.id, p]))
  const deviceNameById = new Map(
    devices.filter((d) => d.id && d.name).map((d) => [d.id, d.name])
  )
  const getRouterName = (deviceId: string, deviceName?: string) =>
    deviceName || deviceNameById.get(deviceId) || deviceId || '-'

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} disalin ke clipboard`)
  }

  return (
    <div className='space-y-4 p-6'>
      {/* ─── Header bar ─── */}
      <div className='flex items-center justify-between'>
        <div>
          <h3 className='text-sm font-semibold text-foreground'>Layanan Internet Aktif</h3>
          <p className='text-xs text-muted-foreground'>
            Kelola akun PPPoE, Hotspot, dan konfigurasi provisi router MikroTik.
          </p>
        </div>
        <Button size='sm' onClick={onAddSubscription} className='h-8 gap-1.5 shadow-xs'>
          <Plus className='h-3.5 w-3.5' />
          Tambah Layanan
        </Button>
      </div>

      {isLoading ? (
        <div className='py-16 text-center text-sm text-muted-foreground animate-pulse'>
          Memuat data layanan...
        </div>
      ) : subscriptions.length === 0 ? (
        <div className='rounded-xl border border-dashed py-14 text-center'>
          <Repeat className='mx-auto h-10 w-10 text-muted-foreground/40' />
          <p className='mt-3 text-sm font-semibold text-foreground'>Belum Ada Layanan Terdaftar</p>
          <p className='mt-1 text-xs text-muted-foreground max-w-xs mx-auto'>
            Pelanggan ini belum memiliki akun langganan internet PPPoE, Hotspot, atau Dedicated.
          </p>
          <Button size='sm' variant='outline' className='mt-4 gap-1.5' onClick={onAddSubscription}>
            <Plus className='h-3.5 w-3.5' />
            Buat Langganan Pertama
          </Button>
        </div>
      ) : (
        <div className='space-y-3.5'>
          {subscriptions.map((sub) => {
            const statusBadge = subscriptionStatusBadge(sub.status)
            const provMeta =
              PROVISION_STATUS_META[sub.provisionStatus as keyof typeof PROVISION_STATUS_META] ||
              PROVISION_STATUS_META.NONE
            const isHotspot = sub.serviceType === 'HOTSPOT'
            const isDedicated = sub.serviceType === 'DEDICATED'
            const plan = planById.get(sub.planId)
            const planName = sub.planName || plan?.name || sub.planId
            const rateLimit =
              sub.pppoeConfig?.rateLimit ||
              sub.hotspotConfig?.rateLimit ||
              sub.rateLimit ||
              formatPlanRate(plan?.bandwidthDownloadKbps, plan?.bandwidthUploadKbps) ||
              null
            const routerName = getRouterName(sub.deviceId, sub.deviceName)
            const effectivePrice =
              sub.customPrice > 0
                ? sub.customPrice
                : sub.planPrice > 0
                ? sub.planPrice
                : (plan?.price ?? 0)

            const staticIP = sub.pppoeConfig?.remoteAddress || sub.hotspotConfig?.ipAddress

            return (
              <div
                key={sub.id}
                className='overflow-hidden rounded-xl border bg-card shadow-xs transition-all hover:border-primary/40'
              >
                {/* Header Kartu */}
                <div className='flex items-start justify-between gap-3 p-4 pb-3 border-b bg-muted/15'>
                  <div className='flex items-center gap-3 min-w-0'>
                    <div
                      className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${
                        isHotspot
                          ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400'
                          : isDedicated
                          ? 'bg-orange-500/10 text-orange-600 dark:text-orange-400'
                          : 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
                      }`}
                    >
                      {isHotspot ? (
                        <Wifi className='h-5 w-5' />
                      ) : isDedicated ? (
                        <Server className='h-5 w-5' />
                      ) : (
                        <Network className='h-5 w-5' />
                      )}
                    </div>
                    <div className='min-w-0'>
                      <div className='flex items-center gap-2'>
                        <span className='font-mono text-sm font-bold truncate text-foreground'>
                          {sub.remoteUsername}
                        </span>
                        <Button
                          size='icon'
                          variant='ghost'
                          className='h-6 w-6 text-muted-foreground hover:text-foreground'
                          onClick={() => copyToClipboard(sub.remoteUsername, 'Username')}
                          title='Salin username'
                        >
                          <Copy className='h-3 w-3' />
                        </Button>
                      </div>
                      <p className='text-xs text-muted-foreground truncate font-medium'>
                        {planName}
                      </p>
                    </div>
                  </div>

                  {/* Status Badges */}
                  <div className='flex shrink-0 flex-col items-end gap-1.5'>
                    <Badge variant='outline' className={`text-[10px] px-2 py-0.5 ${statusBadge.className}`}>
                      {statusBadge.label}
                    </Badge>
                    <Badge variant='outline' className={`text-[10px] px-2 py-0.5 ${provMeta.className}`}>
                      {provMeta.label}
                    </Badge>
                  </div>
                </div>

                {/* Key Metrics Chips */}
                <div className='grid grid-cols-2 sm:grid-cols-4 gap-2.5 p-4 py-3 bg-muted/5 border-b text-xs'>
                  {/* Router */}
                  <div>
                    <span className='text-[10px] text-muted-foreground flex items-center gap-1'>
                      <Router className='h-3 w-3' /> Router
                    </span>
                    <p className='font-semibold mt-0.5 truncate text-foreground'>{routerName}</p>
                  </div>

                  {/* Bandwidth / Rate Limit */}
                  <div>
                    <span className='text-[10px] text-muted-foreground flex items-center gap-1'>
                      <Gauge className='h-3 w-3' /> Kecepatan
                    </span>
                    <p className='font-mono font-bold mt-0.5 text-primary'>
                      {rateLimit ?? '-'}
                    </p>
                  </div>

                  {/* Tarif Bulanan */}
                  <div>
                    <span className='text-[10px] text-muted-foreground'>Tarif Bulanan</span>
                    <p className='font-semibold mt-0.5 text-foreground'>
                      {formatCurrency(effectivePrice)}
                      {sub.customPrice > 0 && (
                        <span className='ml-1 text-[10px] text-amber-500 font-normal'>(khusus)</span>
                      )}
                    </p>
                  </div>

                  {/* Siklus Tagihan */}
                  <div>
                    <span className='text-[10px] text-muted-foreground'>Jatuh Tempo</span>
                    <p className='font-semibold mt-0.5 text-foreground'>
                      Tgl {sub.billingDay || 1} ({sub.billingCycle || 'MONTHLY'})
                    </p>
                  </div>
                </div>

                {/* Detail Jaringan Kompak */}
                <div className='px-4 py-2.5 bg-card text-xs space-y-1.5'>
                  <div className='flex flex-wrap items-center gap-x-6 gap-y-1.5 text-muted-foreground'>
                    {staticIP && (
                      <div className='flex items-center gap-1.5 font-mono'>
                        <span className='text-[10px] font-sans font-medium text-muted-foreground'>Static IP:</span>
                        <span className='font-semibold text-foreground'>{staticIP}</span>
                        <Button
                          size='icon'
                          variant='ghost'
                          className='h-5 w-5 text-muted-foreground hover:text-foreground'
                          onClick={() => copyToClipboard(staticIP, 'IP Address')}
                          title='Salin IP'
                        >
                          <Copy className='h-2.5 w-2.5' />
                        </Button>
                      </div>
                    )}

                    {!isHotspot && sub.pppoeConfig?.callerId && (
                      <div className='flex items-center gap-1.5 font-mono text-[11px]'>
                        <span className='font-sans text-muted-foreground'>Caller ID (MAC ONT):</span>
                        <span className='font-medium text-foreground'>{sub.pppoeConfig.callerId}</span>
                      </div>
                    )}

                    {isHotspot && sub.hotspotConfig?.macAddress && (
                      <div className='flex items-center gap-1.5 font-mono text-[11px]'>
                        <span className='font-sans text-muted-foreground'>Binding MAC:</span>
                        <span className='font-medium text-foreground'>{sub.hotspotConfig.macAddress}</span>
                      </div>
                    )}

                    {isHotspot && sub.hotspotConfig?.limitUptime && (
                      <div className='flex items-center gap-1.5 text-[11px]'>
                        <span className='text-muted-foreground'>Uptime:</span>
                        <span className='font-medium text-foreground'>{sub.hotspotConfig.limitUptime}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Footer Tombol Aksi */}
                <div className='flex items-center justify-end gap-2 px-4 py-2 border-t bg-muted/20'>
                  <Button
                    size='sm'
                    variant='outline'
                    className='h-7 gap-1 text-xs'
                    onClick={() => onEditSubscription(sub)}
                  >
                    <Pencil className='h-3 w-3' />
                    Edit Layanan
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
