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
import { customerStatusBadge } from '../data/constants'
import { PROVISION_STATUS_META, subscriptionStatusBadge } from '@/features/billing/data/constants'
import { useCustomers } from './customers-provider'

interface CustomersDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  customer: Customer | null
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
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
                ID: {customer.customerCode || customer.id}
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
                Overview
              </TabsTrigger>
              <TabsTrigger
                value='subscriptions'
                className='relative h-11 rounded-none border-b-2 border-transparent px-4 pb-3 pt-2 font-medium text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground'
              >
                Langganan ({subscriptions.length})
              </TabsTrigger>
              <TabsTrigger
                value='invoices'
                className='relative h-11 rounded-none border-b-2 border-transparent px-4 pb-3 pt-2 font-medium text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground'
              >
                Tagihan ({invoices.length})
              </TabsTrigger>
            </TabsList>
          </div>

          <ScrollArea className='flex-1 p-6'>
            <TabsContent value='overview' className='m-0 space-y-6'>
              {/* Quick Contacts */}
              <div className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
                <Card className='bg-muted/40'>
                  <CardContent className='flex items-center gap-3 p-4'>
                    <div className='flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-green-500/10 text-green-600 dark:text-green-400'>
                      <MessageCircle className='h-5 w-5' />
                    </div>
                    <div className='min-w-0 flex-1'>
                      <p className='text-xs text-muted-foreground'>WhatsApp / Telepon</p>
                      <p className='font-mono text-sm font-semibold truncate'>{customer.phone || '-'}</p>
                    </div>
                    {customer.phone && (
                      <Button
                        size='sm'
                        variant='ghost'
                        className='h-8 w-8 p-0 text-green-600'
                        asChild
                      >
                        <a
                          href={`https://wa.me/${waPhone}`}
                          target='_blank'
                          rel='noopener noreferrer'
                          title='Chat WhatsApp'
                        >
                          <ExternalLink className='h-4 w-4' />
                        </a>
                      </Button>
                    )}
                  </CardContent>
                </Card>

                <Card className='bg-muted/40'>
                  <CardContent className='flex items-center gap-3 p-4'>
                    <div className='flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400'>
                      <Mail className='h-5 w-5' />
                    </div>
                    <div className='min-w-0 flex-1'>
                      <p className='text-xs text-muted-foreground'>Email</p>
                      <p className='text-sm font-medium truncate'>{customer.email || '-'}</p>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Identity & Portal Info */}
              <Card>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold'>Informasi Akun & Akses</CardTitle>
                </CardHeader>
                <CardContent className='grid grid-cols-2 gap-4 p-4 pt-2 text-sm'>
                  <div>
                    <span className='text-xs text-muted-foreground'>Kode Akses Portal:</span>
                    <div className='mt-1 flex items-center gap-2'>
                      <KeyRound className='h-4 w-4 text-muted-foreground' />
                      <code className='rounded bg-muted px-2 py-0.5 font-mono text-sm font-bold'>
                        {customer.portalAccessCode || '-'}
                      </code>
                    </div>
                  </div>
                  <div>
                    <span className='text-xs text-muted-foreground'>Tanggal Terdaftar:</span>
                    <div className='mt-1 flex items-center gap-2'>
                      <Calendar className='h-4 w-4 text-muted-foreground' />
                      <span>{formatUnixDate(customer.registeredAtUnix || customer.createdAtUnix)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Address & Coordinates */}
              <Card>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold'>Alamat & Lokasi Pemasangan</CardTitle>
                </CardHeader>
                <CardContent className='space-y-3 p-4 pt-2 text-sm'>
                  <div className='flex items-start gap-2'>
                    <Building className='mt-0.5 h-4 w-4 shrink-0 text-muted-foreground' />
                    <span>{customer.address || '-'}</span>
                  </div>
                  {customer.hasCoordinates && customer.latitude !== undefined && customer.longitude !== undefined && (
                    <div className='flex items-center justify-between rounded-md bg-muted/50 p-2.5'>
                      <div className='flex items-center gap-2 font-mono text-xs'>
                        <MapPin className='h-4 w-4 text-rose-500' />
                        <span>
                          {customer.latitude.toFixed(6)}, {customer.longitude.toFixed(6)}
                        </span>
                      </div>
                      <Button size='sm' variant='outline' className='h-7 text-xs' asChild>
                        <a
                          href={`https://www.google.com/maps?q=${customer.latitude},${customer.longitude}`}
                          target='_blank'
                          rel='noopener noreferrer'
                        >
                          Buka di Google Maps
                        </a>
                      </Button>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Notes */}
              {customer.notes && (
                <Card>
                  <CardHeader className='p-4 pb-2'>
                    <CardTitle className='text-sm font-semibold'>Catatan Pelanggan</CardTitle>
                  </CardHeader>
                  <CardContent className='p-4 pt-2 text-sm text-muted-foreground'>
                    {customer.notes}
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            <TabsContent value='subscriptions' className='m-0 space-y-4'>
              <div className='flex items-center justify-between'>
                <p className='text-xs text-muted-foreground'>
                  Semua layanan aktif dan riwayat koneksi pelanggan ini.
                </p>
                <Button size='sm' onClick={handleAddSubscription} className='gap-1'>
                  <Plus className='h-3.5 w-3.5' /> Tambah Langganan
                </Button>
              </div>

              {isLoadingSubs ? (
                <div className='p-8 text-center text-sm text-muted-foreground'>Memuat langganan...</div>
              ) : subscriptions.length === 0 ? (
                <div className='rounded-lg border border-dashed p-8 text-center'>
                  <Repeat className='mx-auto h-8 w-8 text-muted-foreground/50' />
                  <h4 className='mt-2 text-sm font-semibold'>Belum Ada Langganan</h4>
                  <p className='text-xs text-muted-foreground mt-1'>
                    Pelanggan ini belum memiliki akun PPPoE atau Hotspot aktif.
                  </p>
                  <Button size='sm' variant='outline' className='mt-4' onClick={handleAddSubscription}>
                    Buat Langganan Pertama
                  </Button>
                </div>
              ) : (
                <div className='space-y-3'>
                  {subscriptions.map((sub) => {
                    const subStatus = subscriptionStatusBadge(sub.status)
                    const provMeta = PROVISION_STATUS_META[sub.provisionStatus as keyof typeof PROVISION_STATUS_META] || {
                      label: sub.provisionStatus,
                      className: 'bg-muted text-muted-foreground',
                    }
                    const isHotspot = sub.serviceType === 'HOTSPOT'

                    return (
                      <Card key={sub.id} className='relative overflow-hidden border'>
                        <div
                          className={`absolute left-0 top-0 bottom-0 w-1 ${
                            sub.status === 'ACTIVE'
                              ? 'bg-emerald-500'
                              : sub.status === 'ISOLATED'
                              ? 'bg-amber-500'
                              : 'bg-rose-500'
                          }`}
                        />
                        <CardHeader className='p-4 pb-2'>
                          <div className='flex items-center justify-between'>
                            <div className='flex items-center gap-2'>
                              <div className='flex h-7 w-7 items-center justify-center rounded bg-muted'>
                                {isHotspot ? (
                                  <Wifi className='h-4 w-4 text-purple-600 dark:text-purple-400' />
                                ) : (
                                  <Network className='h-4 w-4 text-blue-600 dark:text-blue-400' />
                                )}
                              </div>
                              <div>
                                <CardTitle className='font-mono text-sm font-bold'>
                                  {sub.remoteUsername}
                                </CardTitle>
                                <p className='text-xs text-muted-foreground'>{sub.planName || sub.planId}</p>
                              </div>
                            </div>
                            <div className='flex items-center gap-1.5'>
                              <Badge variant='outline' className={`text-[10px] ${provMeta.className}`}>
                                {provMeta.label}
                              </Badge>
                              <Badge variant='outline' className={`text-[10px] ${subStatus.className}`}>
                                {subStatus.label}
                              </Badge>
                            </div>
                          </div>
                        </CardHeader>
                        <CardContent className='grid grid-cols-2 gap-2 p-4 pt-2 text-xs'>
                          <div>
                            <span className='text-muted-foreground'>Router:</span>{' '}
                            <span className='font-semibold'>{sub.deviceName || sub.deviceId || '-'}</span>
                          </div>
                          <div>
                            <span className='text-muted-foreground'>Rate Limit:</span>{' '}
                            <span className='font-mono'>{sub.rateLimit || sub.routerProfile || '-'}</span>
                          </div>
                          <div>
                            <span className='text-muted-foreground'>Siklus Billing:</span>{' '}
                            <span>Tiap tgl {sub.billingDay}</span>
                          </div>
                          <div>
                            <span className='text-muted-foreground'>IP Remote:</span>{' '}
                            <span className='font-mono'>{sub.pppoeConfig?.remoteAddress || '-'}</span>
                          </div>
                        </CardContent>
                      </Card>
                    )
                  })}
                </div>
              )}
            </TabsContent>

            <TabsContent value='invoices' className='m-0 space-y-4'>
              <p className='text-xs text-muted-foreground'>
                Daftar tagihan dan riwayat pembayaran pelanggan.
              </p>

              {isLoadingInvoices ? (
                <div className='p-8 text-center text-sm text-muted-foreground'>Memuat tagihan...</div>
              ) : invoices.length === 0 ? (
                <div className='rounded-lg border border-dashed p-8 text-center'>
                  <Receipt className='mx-auto h-8 w-8 text-muted-foreground/50' />
                  <h4 className='mt-2 text-sm font-semibold'>Belum Ada Tagihan</h4>
                  <p className='text-xs text-muted-foreground mt-1'>
                    Belum ada invoice yang diterbitkan untuk pelanggan ini.
                  </p>
                </div>
              ) : (
                <div className='space-y-2.5'>
                  {invoices.map((inv) => {
                    const isPaid = inv.status === 'PAID'
                    return (
                      <div
                        key={inv.id}
                        className='flex items-center justify-between rounded-lg border p-3 text-sm'
                      >
                        <div>
                          <div className='flex items-center gap-2'>
                            <span className='font-mono font-semibold'>{inv.invoiceNumber || inv.id}</span>
                            <Badge
                              variant='outline'
                              className={`text-[10px] ${
                                isPaid
                                  ? 'bg-green-500/15 text-green-700 dark:text-green-400 border-green-500/30'
                                  : 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30'
                              }`}
                            >
                              {inv.status}
                            </Badge>
                          </div>
                          <p className='text-xs text-muted-foreground mt-0.5'>
                            Periode: {inv.period || '-'} • Jatuh Tempo: {formatUnixDate(inv.dueDateUnix)}
                          </p>
                        </div>
                        <div className='text-right'>
                          <p className='font-semibold'>{formatCurrency(inv.total)}</p>
                          {inv.paidAtUnix ? (
                            <p className='text-[10px] text-muted-foreground'>
                              Dibayar: {formatUnixDate(inv.paidAtUnix)}
                            </p>
                          ) : null}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </TabsContent>
          </ScrollArea>
        </Tabs>
      </SheetContent>
    </Sheet>
  )
}
