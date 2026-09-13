import {
  ArrowRight,
  Building,
  Copy,
  ExternalLink,
  FileText,
  KeyRound,
  MapPin,
  MessageCircle,
  Phone,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import type { Customer } from '@/gen/v1/customer_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Invoice } from '@/gen/v1/billing_pb'
import { subscriptionStatusBadge } from '@/features/billing/data/constants'
import { invoiceStatusBadge } from '@/features/invoices/data/constants'
import { formatCurrency, formatUnixDate, toWaPhone } from '@/lib/format'

type DetailTab = 'ringkasan' | 'layanan' | 'tagihan'

interface CustomerOverviewTabProps {
  customer: Customer
  subscriptions: Subscription[]
  invoices: Invoice[]
  onOpenTab: (tab: DetailTab) => void
  onOpenInvoice: (invoice: Invoice) => void
}

function InfoRow({
  label,
  value,
  mono,
  onCopy,
}: {
  label: string
  value: string
  mono?: boolean
  onCopy?: () => void
}) {
  return (
    <div className='flex items-center justify-between gap-3 rounded-lg bg-muted/40 px-3 py-2'>
      <div className='min-w-0'>
        <p className='text-[10px] uppercase tracking-wide text-muted-foreground'>
          {label}
        </p>
        <p className={`truncate text-sm ${mono ? 'font-mono' : ''}`}>{value}</p>
      </div>
      {onCopy && (
        <Button
          size='icon'
          variant='ghost'
          className='h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground'
          onClick={onCopy}
          title={`Salin ${label.toLowerCase()}`}
        >
          <Copy className='h-3.5 w-3.5' />
        </Button>
      )}
    </div>
  )
}

export function CustomerOverviewTab({
  customer,
  subscriptions,
  invoices,
  onOpenTab,
  onOpenInvoice,
}: CustomerOverviewTabProps) {
  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} berhasil disalin`)
  }

  const waPhone = toWaPhone(customer.phone)
  const recentSubscriptions = subscriptions.slice(0, 3)
  const recentInvoices = invoices.slice(0, 3)

  return (
    <div className='space-y-4'>
      <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
        {/* Kontak */}
        <Card className='shadow-xs'>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-sm'>
              <Phone className='h-4 w-4 text-emerald-600 dark:text-emerald-400' />
              Kontak & WhatsApp
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-2.5'>
            <InfoRow
              label='Nomor HP / WhatsApp'
              value={customer.phone || '-'}
              mono
              onCopy={customer.phone ? () => copyToClipboard(customer.phone, 'Nomor HP') : undefined}
            />
            <InfoRow
              label='Email'
              value={customer.email || '-'}
              onCopy={customer.email ? () => copyToClipboard(customer.email ?? '', 'Email') : undefined}
            />
            {customer.phone && (
              <Button
                variant='outline'
                size='sm'
                className='h-8 w-full gap-1.5 border-emerald-500/30 text-xs text-emerald-600 hover:bg-emerald-500/10 hover:text-emerald-700'
                asChild
              >
                <a
                  href={`https://wa.me/${waPhone}?text=${encodeURIComponent(`Halo ${customer.name}`)}`}
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  <MessageCircle className='h-3.5 w-3.5' />
                  Chat WhatsApp
                </a>
              </Button>
            )}
          </CardContent>
        </Card>

        {/* Portal */}
        <Card className='border-primary/20 bg-primary/5 shadow-xs'>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-sm'>
              <KeyRound className='h-4 w-4 text-primary' />
              Akses Portal Mandiri
            </CardTitle>
            <CardDescription className='text-[11px]'>
              Kode ini dipakai pelanggan untuk cek tagihan dan pemulihan isolir
              mandiri di portal.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className='flex items-center justify-between gap-3 rounded-lg border bg-background/80 px-3 py-2 shadow-xs'>
              <div>
                <p className='text-[10px] uppercase tracking-wide text-muted-foreground'>
                  Kode Sandi Masuk
                </p>
                <code className='font-mono text-base font-extrabold tracking-wider'>
                  {customer.portalAccessCode || '-'}
                </code>
              </div>
              {customer.portalAccessCode && (
                <Button
                  variant='outline'
                  size='sm'
                  className='h-8 gap-1.5 text-xs'
                  onClick={() =>
                    copyToClipboard(customer.portalAccessCode, 'Kode akses portal')
                  }
                >
                  <Copy className='h-3.5 w-3.5' />
                  Salin
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Alamat */}
        <Card className='shadow-xs'>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center justify-between text-sm'>
              <span className='flex items-center gap-2'>
                <Building className='h-4 w-4 text-muted-foreground' />
                Alamat Pemasangan
              </span>
              {customer.hasCoordinates && (
                <Button variant='outline' size='sm' className='h-7 gap-1.5 text-xs' asChild>
                  <a
                    href={`https://www.google.com/maps?q=${customer.latitude},${customer.longitude}`}
                    target='_blank'
                    rel='noopener noreferrer'
                  >
                    <ExternalLink className='h-3.5 w-3.5 text-rose-500' />
                    Google Maps
                  </a>
                </Button>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-2.5'>
            <p className='text-sm leading-relaxed text-foreground/90'>
              {customer.address || 'Alamat belum diisi.'}
            </p>
            {customer.hasCoordinates && (
              <div className='flex items-center gap-2 rounded-lg bg-muted/40 px-3 py-2 font-mono text-xs text-muted-foreground'>
                <MapPin className='h-4 w-4 shrink-0 text-rose-500' />
                {customer.latitude.toFixed(6)}, {customer.longitude.toFixed(6)}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Catatan */}
        <Card className='shadow-xs'>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-sm'>
              <FileText className='h-4 w-4 text-muted-foreground' />
              Catatan Internal
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className='text-sm leading-relaxed text-muted-foreground'>
              {customer.notes || 'Belum ada catatan internal.'}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
        {/* Layanan terbaru */}
        <Card className='shadow-xs'>
          <CardHeader className='flex flex-row items-center justify-between pb-3'>
            <CardTitle className='text-sm'>Layanan Terbaru</CardTitle>
            <Button
              variant='ghost'
              size='sm'
              className='h-7 gap-1 text-xs'
              onClick={() => onOpenTab('layanan')}
            >
              Lihat Semua
              <ArrowRight className='h-3.5 w-3.5' />
            </Button>
          </CardHeader>
          <CardContent className='space-y-3'>
            {recentSubscriptions.length === 0 ? (
              <p className='text-sm text-muted-foreground'>Belum ada layanan.</p>
            ) : (
              recentSubscriptions.map((sub) => {
                const badge = subscriptionStatusBadge(sub.status)
                return (
                  <div key={sub.id} className='flex items-center justify-between gap-3'>
                    <div className='min-w-0'>
                      <p className='truncate text-sm font-semibold'>
                        {sub.planName || sub.planId}
                      </p>
                      <p className='truncate font-mono text-xs text-muted-foreground'>
                        {sub.remoteUsername}
                      </p>
                    </div>
                    <Badge variant='outline' className={`shrink-0 text-[10px] ${badge.className}`}>
                      {badge.label}
                    </Badge>
                  </div>
                )
              })
            )}
          </CardContent>
        </Card>

        {/* Tagihan terbaru */}
        <Card className='shadow-xs'>
          <CardHeader className='flex flex-row items-center justify-between pb-3'>
            <CardTitle className='text-sm'>Tagihan Terbaru</CardTitle>
            <Button
              variant='ghost'
              size='sm'
              className='h-7 gap-1 text-xs'
              onClick={() => onOpenTab('tagihan')}
            >
              Lihat Semua
              <ArrowRight className='h-3.5 w-3.5' />
            </Button>
          </CardHeader>
          <CardContent className='space-y-3'>
            {recentInvoices.length === 0 ? (
              <p className='text-sm text-muted-foreground'>Belum ada tagihan.</p>
            ) : (
              recentInvoices.map((inv, index) => {
                const badge = invoiceStatusBadge(inv.status)
                return (
                  <div key={inv.id}>
                    {index > 0 && <Separator className='mb-3' />}
                    <button
                      type='button'
                      onClick={() => onOpenInvoice(inv)}
                      className='flex w-full items-center justify-between gap-3 text-left'
                    >
                      <div className='min-w-0'>
                        <p className='truncate font-mono text-sm font-semibold'>
                          {inv.invoiceNumber || inv.id}
                        </p>
                        <p className='text-xs text-muted-foreground'>
                          {inv.period || '-'} · Jatuh tempo{' '}
                          {formatUnixDate(inv.dueDateUnix)}
                        </p>
                      </div>
                      <div className='shrink-0 text-right'>
                        <p className='font-mono text-sm font-bold'>
                          {formatCurrency(inv.total)}
                        </p>
                        <Badge
                          variant='outline'
                          className={`text-[10px] ${badge.className}`}
                        >
                          {badge.label}
                        </Badge>
                      </div>
                    </button>
                  </div>
                )
              })
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
