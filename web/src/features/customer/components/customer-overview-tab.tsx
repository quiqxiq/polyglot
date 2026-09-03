import {
  Building,
  Copy,
  ExternalLink,
  KeyRound,
  Mail,
  MapPin,
  MessageCircle,
  Phone,
  FileText,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { toast } from 'sonner'
import type { Customer } from '@/gen/v1/customer_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Invoice } from '@/gen/v1/billing_pb'
import { customerStatusBadge } from '../data/constants'

interface CustomerOverviewTabProps {
  customer: Customer
  subscriptions: Subscription[]
  invoices: Invoice[]
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

export function CustomerOverviewTab({
  customer,
  subscriptions,
  invoices,
}: CustomerOverviewTabProps) {
  const statusMeta = customerStatusBadge(customer.status)
  const activeSubs = subscriptions.filter((s) => s.status === 'ACTIVE').length
  const unpaidInvoices = invoices.filter((i) => i.status !== 'PAID')
  const unpaidCount = unpaidInvoices.length
  const unpaidAmount = unpaidInvoices.reduce(
    (sum, inv) => sum + (inv.total - inv.paidAmount),
    0
  )

  const cleanPhone = customer.phone.replace(/\D/g, '')
  const waPhone = cleanPhone.startsWith('0') ? `62${cleanPhone.slice(1)}` : cleanPhone

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} berhasil disalin ke clipboard`)
  }

  return (
    <div className='space-y-5 p-6'>
      {/* ─── Executive Metric Cards ─── */}
      <div className='grid grid-cols-1 gap-3 sm:grid-cols-3'>
        {/* Layanan Aktif */}
        <div className='relative overflow-hidden rounded-xl border bg-card p-4 shadow-xs transition-all hover:shadow-xs'>
          <div className='flex items-center justify-between'>
            <span className='text-xs font-medium text-muted-foreground'>Layanan Aktif</span>
            <div className='h-2 w-2 rounded-full bg-emerald-500 animate-pulse' />
          </div>
          <div className='mt-2 flex items-baseline gap-2'>
            <span className='text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400'>
              {activeSubs}
            </span>
            <span className='text-xs text-muted-foreground'>dari {subscriptions.length} total</span>
          </div>
        </div>

        {/* Tagihan Tertunggak */}
        <div className='relative overflow-hidden rounded-xl border bg-card p-4 shadow-xs transition-all hover:shadow-xs'>
          <div className='flex items-center justify-between'>
            <span className='text-xs font-medium text-muted-foreground'>Tagihan Tertunggak</span>
            {unpaidCount > 0 && (
              <Badge variant='outline' className='border-amber-500/30 bg-amber-500/10 text-[10px] text-amber-600 dark:text-amber-400'>
                {unpaidCount} Faktur
              </Badge>
            )}
          </div>
          <div className='mt-2'>
            <span className={`text-xl font-bold tracking-tight font-mono ${unpaidCount > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-foreground'}`}>
              {unpaidCount > 0 ? formatCurrency(unpaidAmount) : 'Rp 0'}
            </span>
            <p className='mt-0.5 text-[11px] text-muted-foreground'>
              {unpaidCount > 0 ? 'Perlu pelunasan segera' : 'Semua tagihan lunas'}
            </p>
          </div>
        </div>

        {/* Status Pelanggan */}
        <div className='relative overflow-hidden rounded-xl border bg-card p-4 shadow-xs transition-all hover:shadow-xs'>
          <div className='flex items-center justify-between'>
            <span className='text-xs font-medium text-muted-foreground'>Status Akun</span>
            <Badge variant='outline' className={`text-[10px] ${statusMeta.className}`}>
              {statusMeta.label}
            </Badge>
          </div>
          <div className='mt-2'>
            <span className='text-xs text-muted-foreground'>Terdaftar sejak:</span>
            <p className='text-sm font-semibold mt-0.5'>
              {formatUnixDate(customer.registeredAtUnix || customer.createdAtUnix) ?? '-'}
            </p>
          </div>
        </div>
      </div>

      {/* ─── Kontak & Akses Portal ─── */}
      <div className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
        {/* WhatsApp & Telepon */}
        <Card className='shadow-xs'>
          <CardContent className='p-4 space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2 text-xs font-semibold text-muted-foreground'>
                <Phone className='h-4 w-4 text-emerald-600 dark:text-emerald-400' />
                <span>Kontak & WhatsApp</span>
              </div>
              {customer.phone && (
                <Button
                  size='sm'
                  variant='outline'
                  className='h-7 gap-1 text-[11px] text-emerald-600 hover:text-emerald-700 hover:bg-emerald-500/10 border-emerald-500/30'
                  asChild
                >
                  <a
                    href={`https://wa.me/${waPhone}?text=Halo%20${encodeURIComponent(customer.name)}`}
                    target='_blank'
                    rel='noopener noreferrer'
                  >
                    <MessageCircle className='h-3.5 w-3.5' />
                    Chat WA
                  </a>
                </Button>
              )}
            </div>

            <div className='flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2'>
              <span className='font-mono text-sm font-bold'>{customer.phone || '-'}</span>
              {customer.phone && (
                <Button
                  size='icon'
                  variant='ghost'
                  className='h-7 w-7 text-muted-foreground hover:text-foreground'
                  onClick={() => copyToClipboard(customer.phone, 'Nomor telepon')}
                  title='Salin nomor'
                >
                  <Copy className='h-3.5 w-3.5' />
                </Button>
              )}
            </div>

            {customer.email && (
              <div className='flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2'>
                <div className='flex items-center gap-1.5 min-w-0'>
                  <Mail className='h-3.5 w-3.5 text-muted-foreground shrink-0' />
                  <span className='text-xs truncate'>{customer.email}</span>
                </div>
                <Button
                  size='icon'
                  variant='ghost'
                  className='h-7 w-7 text-muted-foreground hover:text-foreground shrink-0'
                  onClick={() => copyToClipboard(customer.email, 'Email')}
                  title='Salin email'
                >
                  <Copy className='h-3.5 w-3.5' />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Kode Akses Portal Mandiri */}
        <Card className='shadow-xs border-primary/20 bg-primary/5'>
          <CardContent className='p-4 space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2 text-xs font-semibold text-primary'>
                <KeyRound className='h-4 w-4' />
                <span>Akses Portal Mandiri</span>
              </div>
              <Badge variant='outline' className='text-[10px] border-primary/30 text-primary bg-primary/10'>
                24 Jam Self-Service
              </Badge>
            </div>

            <div className='flex items-center justify-between rounded-lg bg-background/80 border px-3 py-2 shadow-xs'>
              <div>
                <p className='text-[10px] text-muted-foreground'>Kode Sandi Masuk</p>
                <code className='font-mono text-base font-extrabold tracking-wider text-foreground'>
                  {customer.portalAccessCode || '-'}
                </code>
              </div>
              {customer.portalAccessCode && (
                <Button
                  size='sm'
                  variant='outline'
                  className='h-8 gap-1.5 text-xs font-medium'
                  onClick={() => copyToClipboard(customer.portalAccessCode, 'Kode akses portal')}
                >
                  <Copy className='h-3.5 w-3.5' />
                  Salin
                </Button>
              )}
            </div>

            <p className='text-[11px] text-muted-foreground leading-relaxed'>
              Pelanggan dapat mengecek tagihan dan pemulihan isolir mandiri dengan memasukkan kode akses ini di portal pelanggan.
            </p>
          </CardContent>
        </Card>
      </div>

      {/* ─── Alamat Pemasangan & Geolokasi ─── */}
      {(customer.address || customer.hasCoordinates) && (
        <Card className='shadow-xs'>
          <CardContent className='p-4 space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2 text-xs font-semibold text-muted-foreground'>
                <Building className='h-4 w-4' />
                <span>Alamat Pemasangan</span>
              </div>
              {customer.hasCoordinates && customer.latitude !== undefined && customer.longitude !== undefined && (
                <Button size='sm' variant='outline' className='h-7 gap-1.5 text-xs' asChild>
                  <a
                    href={`https://www.google.com/maps?q=${customer.latitude},${customer.longitude}`}
                    target='_blank'
                    rel='noopener noreferrer'
                  >
                    <ExternalLink className='h-3.5 w-3.5 text-rose-500' />
                    Buka Google Maps
                  </a>
                </Button>
              )}
            </div>

            {customer.address && (
              <p className='text-sm leading-relaxed text-foreground/90 pl-1'>
                {customer.address}
              </p>
            )}

            {customer.hasCoordinates && customer.latitude !== undefined && customer.longitude !== undefined && (
              <div className='flex items-center gap-2 rounded-lg bg-muted/40 px-3 py-2 text-xs font-mono text-muted-foreground'>
                <MapPin className='h-4 w-4 text-rose-500 shrink-0' />
                <span>
                  {customer.latitude.toFixed(6)}, {customer.longitude.toFixed(6)}
                </span>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* ─── Catatan Tambahan ─── */}
      {customer.notes && (
        <Card className='shadow-xs'>
          <CardContent className='p-4 space-y-1.5'>
            <div className='flex items-center gap-2 text-xs font-semibold text-muted-foreground'>
              <FileText className='h-4 w-4' />
              <span>Catatan Internal</span>
            </div>
            <p className='text-xs leading-relaxed text-muted-foreground pl-1'>
              {customer.notes}
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
