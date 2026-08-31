import {
  Building,
  Calendar,
  ExternalLink,
  KeyRound,
  Mail,
  MapPin,
  MessageCircle,
  Plus,
  Repeat,
  Receipt,
  Wifi,
  Network,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { Customer } from '@/gen/v1/customer_pb'
import { useSubscriptionsQuery, useInvoicesQuery } from '@/features/billing/api/use-billing'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { customerStatusBadge } from '../data/constants'
import { subscriptionStatusBadge, PROVISION_STATUS_META } from '@/features/billing/data/constants'
import { useCustomers } from './customers-provider'

interface CustomersDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  customer: Customer | null
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return null
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
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

export function CustomersDetailSheet({
  open,
  onOpenChange,
  customer,
}: CustomersDetailSheetProps) {
  const { setOpen, setCurrentRow } = useCustomers()
  const customerId = customer?.id || ''
  const isEnabled = open && Boolean(customerId)

  const { data: subscriptions = [], isLoading: isLoadingSubs } =
    useSubscriptionsQuery(customerId, { enabled: isEnabled })
  const { data: invoices = [], isLoading: isLoadingInvoices } =
    useInvoicesQuery(customerId, '', { enabled: isEnabled })
  const { data: plans = [] } = usePlansQuery(false)
  const { data: devices = [] } = useDevicesQuery()

  if (!customer) return null

  const statusMeta = customerStatusBadge(customer.status)
  const cleanPhone = customer.phone.replace(/\D/g, '')
  const waPhone = cleanPhone.startsWith('0') ? `62${cleanPhone.slice(1)}` : cleanPhone

  const handleAddSubscription = () => {
    setCurrentRow(customer)
    setOpen('create-subscription')
  }

  const handleEdit = () => {
    setCurrentRow(customer)
    setOpen('update')
  }

  const activeCount = subscriptions.filter((s) => s.status === 'ACTIVE').length
  const unpaidCount = invoices.filter((i) => i.status !== 'PAID').length

  /** Lookup map */
  const planById = new Map(plans.map(p => [p.id, p]))
  const deviceNameById = new Map(devices.filter(d => d.id && d.name).map(d => [d.id, d.name]))
  const getRouterName = (deviceId: string, deviceName?: string) =>
    deviceName || deviceNameById.get(deviceId) || deviceId || '-'

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='flex w-full flex-col sm:max-w-xl md:max-w-2xl p-0'>
        <SheetHeader className='border-b p-6 pb-4'>
          <div className='flex items-start justify-between gap-4'>
            <div>
              <div className='flex items-center gap-2'>
                <SheetTitle className='text-xl font-bold'>{customer.name}</SheetTitle>
                <Badge variant='outline' className={`text-xs ${statusMeta.className}`}>
                  {statusMeta.label}
                </Badge>
              </div>
              <SheetDescription className='mt-1 font-mono text-xs text-muted-foreground'>
                {customer.customerCode || customer.id}
              </SheetDescription>
            </div>
            <Button size='sm' variant='outline' onClick={handleEdit}>
              Edit Profil
            </Button>
          </div>
        </SheetHeader>

        <Tabs defaultValue='overview' className='flex flex-1 flex-col overflow-hidden'>
          <div className='border-b px-6'>
            <TabsList className='h-11 w-full justify-start rounded-none bg-transparent p-0'>
              <TabsTrigger
                value='overview'
                className='relative h-11 rounded-none border-b-2 border-transparent px-4 pb-3 pt-2 font-medium text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground'
              >
                Info
              </TabsTrigger>
              <TabsTrigger
                value='subscriptions'
                className='relative h-11 rounded-none border-b-2 border-transparent px-4 pb-3 pt-2 font-medium text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground'
              >
                Langganan
                {subscriptions.length > 0 && (
                  <span className='ml-1.5 rounded-full bg-primary/15 px-1.5 py-0.5 text-[10px] font-semibold text-primary'>
                    {subscriptions.length}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger
                value='invoices'
                className='relative h-11 rounded-none border-b-2 border-transparent px-4 pb-3 pt-2 font-medium text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground'
              >
                Tagihan
                {unpaidCount > 0 && (
                  <span className='ml-1.5 rounded-full bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-semibold text-amber-600 dark:text-amber-400'>
                    {unpaidCount}
                  </span>
                )}
              </TabsTrigger>
            </TabsList>
          </div>

          <ScrollArea className='flex-1'>
            {/* ─── Tab Info ─── */}
            <TabsContent value='overview' className='m-0 p-6 space-y-5'>
              {/* Stats row */}
              <div className='grid grid-cols-3 gap-3'>
                <div className='rounded-xl border bg-card p-3 text-center'>
                  <p className='text-xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400'>{activeCount}</p>
                  <p className='mt-0.5 text-[11px] text-muted-foreground'>Aktif</p>
                </div>
                <div className='rounded-xl border bg-card p-3 text-center'>
                  <p className={`text-xl font-bold tabular-nums ${unpaidCount > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-muted-foreground'}`}>
                    {unpaidCount}
                  </p>
                  <p className='mt-0.5 text-[11px] text-muted-foreground'>Belum Bayar</p>
                </div>
                <div className='rounded-xl border bg-card p-3 text-center'>
                  <p className='text-xl font-bold tabular-nums'>{subscriptions.length}</p>
                  <p className='mt-0.5 text-[11px] text-muted-foreground'>Total</p>
                </div>
              </div>

              {/* Contacts */}
              <div className='space-y-2'>
                {customer.phone && (
                  <div className='flex items-center gap-3 rounded-lg border bg-card px-3 py-2.5'>
                    <div className='flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-green-500/10 text-green-600 dark:text-green-400'>
                      <MessageCircle className='h-4 w-4' />
                    </div>
                    <div className='min-w-0 flex-1'>
                      <p className='text-[10px] text-muted-foreground'>WhatsApp / Telepon</p>
                      <p className='font-mono text-sm font-semibold'>{customer.phone}</p>
                    </div>
                    <Button size='sm' variant='ghost' className='h-7 w-7 p-0 text-green-600' asChild>
                      <a href={`https://wa.me/${waPhone}`} target='_blank' rel='noopener noreferrer'>
                        <ExternalLink className='h-3.5 w-3.5' />
                      </a>
                    </Button>
                  </div>
                )}
                {customer.email && (
                  <div className='flex items-center gap-3 rounded-lg border bg-card px-3 py-2.5'>
                    <div className='flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-blue-500/10 text-blue-600 dark:text-blue-400'>
                      <Mail className='h-4 w-4' />
                    </div>
                    <div className='min-w-0 flex-1'>
                      <p className='text-[10px] text-muted-foreground'>Email</p>
                      <p className='text-sm truncate'>{customer.email}</p>
                    </div>
                  </div>
                )}
              </div>

              {/* Account info */}
              <Card>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold'>Akun & Akses</CardTitle>
                </CardHeader>
                <CardContent className='grid grid-cols-2 gap-4 p-4 pt-2 text-sm'>
                  <div>
                    <p className='text-[11px] text-muted-foreground mb-1'>Kode Akses Portal</p>
                    <div className='flex items-center gap-1.5'>
                      <KeyRound className='h-3.5 w-3.5 text-muted-foreground' />
                      <code className='rounded bg-muted px-1.5 py-0.5 font-mono text-sm font-bold'>
                        {customer.portalAccessCode || '-'}
                      </code>
                    </div>
                  </div>
                  <div>
                    <p className='text-[11px] text-muted-foreground mb-1'>Terdaftar</p>
                    <div className='flex items-center gap-1.5'>
                      <Calendar className='h-3.5 w-3.5 text-muted-foreground' />
                      <span className='text-sm'>
                        {formatUnixDate(customer.registeredAtUnix || customer.createdAtUnix) ?? '-'}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Address */}
              {(customer.address || customer.hasCoordinates) && (
                <Card>
                  <CardHeader className='p-4 pb-2'>
                    <CardTitle className='text-sm font-semibold'>Alamat Pemasangan</CardTitle>
                  </CardHeader>
                  <CardContent className='space-y-2.5 p-4 pt-2 text-sm'>
                    {customer.address && (
                      <div className='flex items-start gap-2 text-sm'>
                        <Building className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                        <span>{customer.address}</span>
                      </div>
                    )}
                    {customer.hasCoordinates && customer.latitude !== undefined && customer.longitude !== undefined && (
                      <div className='flex items-center justify-between rounded-md bg-muted/50 px-2.5 py-2'>
                        <div className='flex items-center gap-1.5 font-mono text-xs'>
                          <MapPin className='h-3.5 w-3.5 text-rose-500' />
                          <span>{customer.latitude.toFixed(6)}, {customer.longitude.toFixed(6)}</span>
                        </div>
                        <Button size='sm' variant='outline' className='h-6 text-[11px]' asChild>
                          <a
                            href={`https://www.google.com/maps?q=${customer.latitude},${customer.longitude}`}
                            target='_blank'
                            rel='noopener noreferrer'
                          >
                            Maps
                          </a>
                        </Button>
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}

              {customer.notes && (
                <Card>
                  <CardHeader className='p-4 pb-2'>
                    <CardTitle className='text-sm font-semibold'>Catatan</CardTitle>
                  </CardHeader>
                  <CardContent className='p-4 pt-2 text-sm text-muted-foreground'>
                    {customer.notes}
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            {/* ─── Tab Langganan ─── */}
            <TabsContent value='subscriptions' className='m-0 p-5 space-y-3'>
              <div className='flex items-center justify-between'>
                <p className='text-xs text-muted-foreground'>
                  {subscriptions.length} layanan
                </p>
                <Button size='sm' onClick={handleAddSubscription} className='h-8 gap-1.5'>
                  <Plus className='h-3.5 w-3.5' />
                  Tambah
                </Button>
              </div>

              {isLoadingSubs ? (
                <div className='py-12 text-center text-sm text-muted-foreground'>Memuat...</div>
              ) : subscriptions.length === 0 ? (
                <div className='rounded-xl border border-dashed py-12 text-center'>
                  <Repeat className='mx-auto h-8 w-8 text-muted-foreground/40' />
                  <p className='mt-3 text-sm font-medium'>Belum ada langganan</p>
                  <p className='mt-1 text-xs text-muted-foreground'>Belum ada akun PPPoE atau Hotspot.</p>
                  <Button size='sm' variant='outline' className='mt-4' onClick={handleAddSubscription}>
                    Buat Langganan
                  </Button>
                </div>
              ) : (
                subscriptions.map((sub) => {
                  const statusBadge = subscriptionStatusBadge(sub.status)
                  const provMeta =
                    PROVISION_STATUS_META[sub.provisionStatus as keyof typeof PROVISION_STATUS_META] ||
                    PROVISION_STATUS_META.NONE
                  const isHotspot = sub.serviceType === 'HOTSPOT'
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
                    sub.customPrice > 0 ? sub.customPrice : (sub.planPrice > 0 ? sub.planPrice : (plan?.price ?? 0))

                  const statusColor =
                    sub.status === 'ACTIVE'
                      ? 'border-l-emerald-500'
                      : sub.status === 'ISOLATED'
                      ? 'border-l-amber-500'
                      : sub.status === 'SUSPENDED'
                      ? 'border-l-rose-400'
                      : 'border-l-slate-400'

                  return (
                    <div
                      key={sub.id}
                      className={`rounded-xl border border-l-4 bg-card ${statusColor} overflow-hidden`}
                    >
                      {/* Header baris */}
                      <div className='flex items-start justify-between gap-2 px-4 pt-3 pb-2'>
                        <div className='flex items-center gap-2.5 min-w-0'>
                          <div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${isHotspot ? 'bg-purple-500/10' : 'bg-blue-500/10'}`}>
                            {isHotspot
                              ? <Wifi className='h-4 w-4 text-purple-600 dark:text-purple-400' />
                              : <Network className='h-4 w-4 text-blue-600 dark:text-blue-400' />
                            }
                          </div>
                          <div className='min-w-0'>
                            <p className='font-mono text-sm font-bold leading-tight truncate'>
                              {sub.remoteUsername}
                            </p>
                            <p className='text-[11px] text-muted-foreground truncate'>
                              {planName}
                            </p>
                          </div>
                        </div>
                        <div className='flex shrink-0 flex-col items-end gap-1'>
                          <Badge variant='outline' className={`h-5 text-[10px] px-1.5 ${statusBadge.className}`}>
                            {statusBadge.label}
                          </Badge>
                          <Badge variant='outline' className={`h-5 text-[10px] px-1.5 ${provMeta.className}`}>
                            {provMeta.label}
                          </Badge>
                        </div>
                      </div>

                      {/* Detail grid */}
                      <div className='border-t bg-muted/20 px-4 py-2.5'>
                        <dl className='grid grid-cols-2 gap-x-6 gap-y-1.5'>
                          {/* Paket & Tipe */}
                          <div>
                            <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>Tipe Layanan</dt>
                            <dd className='mt-0.5 text-xs font-semibold'>
                              {isHotspot ? 'Hotspot' : sub.serviceType === 'DEDICATED' ? 'Dedicated' : 'PPPoE'}
                            </dd>
                          </div>

                          {/* Router */}
                          <div>
                            <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>Router</dt>
                            <dd className='mt-0.5 text-xs font-semibold truncate'>{routerName}</dd>
                          </div>

                          {/* Rate Limit */}
                          <div>
                            <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>Rate Limit</dt>
                            <dd className='mt-0.5 font-mono text-xs font-bold text-primary'>
                              {rateLimit ?? <span className='text-muted-foreground font-normal'>-</span>}
                            </dd>
                          </div>

                          {/* Profil Router */}
                          {(sub.pppoeConfig?.routerProfile || sub.hotspotConfig?.routerProfile || sub.routerProfile) && (
                            <div>
                              <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>Profil Router</dt>
                              <dd className='mt-0.5 font-mono text-xs'>
                                {sub.pppoeConfig?.routerProfile || sub.hotspotConfig?.routerProfile || sub.routerProfile}
                              </dd>
                            </div>
                          )}

                          {/* Harga */}
                          {effectivePrice > 0 && (
                            <div>
                              <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>
                                Tagihan Bulanan {sub.customPrice > 0 && <span className='text-amber-500'>(custom)</span>}
                              </dt>
                              <dd className='mt-0.5 text-xs font-bold'>{formatCurrency(effectivePrice)}</dd>
                            </div>
                          )}

                          {/* Siklus billing */}
                          <div>
                            <dt className='text-[10px] font-medium uppercase tracking-wide text-muted-foreground'>Siklus Tagihan</dt>
                            <dd className='mt-0.5 text-xs'>
                              {sub.billingCycle || '-'} · tgl {sub.billingDay}
                            </dd>
                          </div>
                        </dl>
                      </div>

                      {/* PPPoE extra */}
                      {!isHotspot && (sub.pppoeConfig?.remoteAddress || sub.pppoeConfig?.localAddress || sub.pppoeConfig?.callerId) && (
                        <div className='border-t px-4 py-2'>
                          <p className='mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground'>Konfigurasi PPPoE</p>
                          <dl className='grid grid-cols-2 gap-x-6 gap-y-1'>
                            {sub.pppoeConfig?.remoteAddress && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Static IP</dt>
                                <dd className='font-mono text-xs font-semibold'>{sub.pppoeConfig.remoteAddress}</dd>
                              </div>
                            )}
                            {sub.pppoeConfig?.localAddress && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Gateway</dt>
                                <dd className='font-mono text-xs'>{sub.pppoeConfig.localAddress}</dd>
                              </div>
                            )}
                            {sub.pppoeConfig?.callerId && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Caller ID</dt>
                                <dd className='font-mono text-xs'>{sub.pppoeConfig.callerId}</dd>
                              </div>
                            )}
                          </dl>
                        </div>
                      )}

                      {/* Hotspot extra */}
                      {isHotspot && (sub.hotspotConfig?.server || sub.hotspotConfig?.ipAddress || sub.hotspotConfig?.macAddress) && (
                        <div className='border-t px-4 py-2'>
                          <p className='mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground'>Konfigurasi Hotspot</p>
                          <dl className='grid grid-cols-2 gap-x-6 gap-y-1'>
                            {sub.hotspotConfig?.server && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Server</dt>
                                <dd className='font-mono text-xs'>{sub.hotspotConfig.server}</dd>
                              </div>
                            )}
                            {sub.hotspotConfig?.ipAddress && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Static IP</dt>
                                <dd className='font-mono text-xs font-semibold'>{sub.hotspotConfig.ipAddress}</dd>
                              </div>
                            )}
                            {sub.hotspotConfig?.macAddress && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>MAC</dt>
                                <dd className='font-mono text-xs'>{sub.hotspotConfig.macAddress}</dd>
                              </div>
                            )}
                            {sub.hotspotConfig?.limitUptime && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Limit Uptime</dt>
                                <dd className='font-mono text-xs'>{sub.hotspotConfig.limitUptime}</dd>
                              </div>
                            )}
                            {sub.hotspotConfig?.limitBytes && (
                              <div>
                                <dt className='text-[10px] text-muted-foreground'>Limit Bytes</dt>
                                <dd className='font-mono text-xs'>{sub.hotspotConfig.limitBytes}</dd>
                              </div>
                            )}
                          </dl>
                        </div>
                      )}

                      {/* Tanggal */}
                      {(Number(sub.startDateUnix) > 0 || Number(sub.endDateUnix) > 0) && (
                        <div className='border-t px-4 py-2 flex gap-6'>
                          {Number(sub.startDateUnix) > 0 && (
                            <div>
                              <p className='text-[10px] text-muted-foreground'>Mulai</p>
                              <p className='text-xs font-medium'>{formatUnixDate(sub.startDateUnix)}</p>
                            </div>
                          )}
                          {Number(sub.endDateUnix) > 0 && (
                            <div>
                              <p className='text-[10px] text-muted-foreground'>Berakhir</p>
                              <p className='text-xs font-medium'>{formatUnixDate(sub.endDateUnix)}</p>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )
                })
              )}
            </TabsContent>

            {/* ─── Tab Tagihan ─── */}
            <TabsContent value='invoices' className='m-0 p-5 space-y-2.5'>
              <p className='text-xs text-muted-foreground'>Riwayat tagihan dan pembayaran.</p>

              {isLoadingInvoices ? (
                <div className='py-12 text-center text-sm text-muted-foreground'>Memuat...</div>
              ) : invoices.length === 0 ? (
                <div className='rounded-xl border border-dashed py-12 text-center'>
                  <Receipt className='mx-auto h-8 w-8 text-muted-foreground/40' />
                  <p className='mt-3 text-sm font-medium'>Belum ada tagihan</p>
                  <p className='mt-1 text-xs text-muted-foreground'>Invoice belum diterbitkan untuk pelanggan ini.</p>
                </div>
              ) : (
                invoices.map((inv) => {
                  const isPaid = inv.status === 'PAID'
                  const isOverdue = inv.status === 'OVERDUE'
                  return (
                    <div
                      key={inv.id}
                      className='flex items-center justify-between rounded-xl border bg-card px-4 py-3'
                    >
                      <div className='min-w-0'>
                        <div className='flex items-center gap-2'>
                          <span className='font-mono text-sm font-semibold'>
                            {inv.invoiceNumber || inv.id}
                          </span>
                          <Badge
                            variant='outline'
                            className={`h-5 text-[10px] px-1.5 ${
                              isPaid
                                ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
                                : isOverdue
                                ? 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30'
                                : 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30'
                            }`}
                          >
                            {inv.status === 'PAID' ? 'Lunas' : inv.status === 'OVERDUE' ? 'Jatuh Tempo' : 'Belum Bayar'}
                          </Badge>
                        </div>
                        <p className='mt-0.5 text-[11px] text-muted-foreground'>
                          {inv.period || '-'}
                          {formatUnixDate(inv.dueDateUnix) && ` · Jatuh tempo ${formatUnixDate(inv.dueDateUnix)}`}
                        </p>
                      </div>
                      <div className='shrink-0 text-right'>
                        <p className='text-sm font-bold'>{new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(inv.total)}</p>
                        {inv.paidAtUnix ? (
                          <p className='text-[10px] text-emerald-600 dark:text-emerald-400'>
                            Dibayar {formatUnixDate(inv.paidAtUnix)}
                          </p>
                        ) : null}
                      </div>
                    </div>
                  )
                })
              )}
            </TabsContent>
          </ScrollArea>
        </Tabs>
      </SheetContent>
    </Sheet>
  )
}
