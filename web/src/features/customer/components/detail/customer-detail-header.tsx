import { ArrowLeft, Copy, CreditCard, MessageCircle, Pencil, Plus } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { toast } from 'sonner'
import type { Customer } from '@/gen/v1/customer_pb'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { formatCurrency, toWaPhone } from '@/lib/format'
import { customerStatusBadge } from '../../data/constants'

interface CustomerDetailHeaderProps {
  customer: Customer
  invoices: Invoice[]
  onEdit: () => void
  onAddSubscription: () => void
  onPay: (invoice: Invoice) => void
}

function getInitials(name: string) {
  const parts = name.trim().split(/\s+/).slice(0, 2)
  const initials = parts.map((part) => part.charAt(0).toUpperCase()).join('')
  return initials || '?'
}

export function CustomerDetailHeader({
  customer,
  invoices,
  onEdit,
  onAddSubscription,
  onPay,
}: CustomerDetailHeaderProps) {
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'customer:manage')
  const canBilling = canPermission(permissions, 'billing:manage')

  const statusMeta = customerStatusBadge(customer.status)
  const unpaidInvoices = invoices.filter((inv) => inv.status !== 'PAID')
  const outstanding = unpaidInvoices.reduce(
    (sum, inv) => sum + Math.max(0, inv.total - inv.paidAmount),
    0
  )
  const dueInvoice =
    [...unpaidInvoices]
      .filter((inv) => Number(inv.dueDateUnix) > 0)
      .sort((a, b) => Number(a.dueDateUnix) - Number(b.dueDateUnix))[0] ??
    unpaidInvoices[0]

  const waPhone = toWaPhone(customer.phone)
  const waMessage =
    outstanding > 0
      ? `Halo ${customer.name}, tagihan internet Anda sebesar ${formatCurrency(outstanding)} belum lunas. Mohon segera melakukan pembayaran. Terima kasih.`
      : `Halo ${customer.name}, ada yang bisa kami bantu terkait layanan internet Anda?`

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} berhasil disalin`)
  }

  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-2 text-sm text-muted-foreground'>
        <Link
          to='/customers'
          className='flex items-center gap-1 transition-colors hover:text-foreground'
        >
          <ArrowLeft className='h-4 w-4' />
          Kembali ke Pelanggan
        </Link>
        <span>/</span>
        <span className='font-medium text-foreground'>Detail Pelanggan</span>
      </div>

      <Card>
        <CardContent className='flex flex-col gap-4 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5'>
          <div className='flex min-w-0 items-center gap-3.5'>
            <div className='flex size-12 shrink-0 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-base font-bold text-primary'>
              {getInitials(customer.name)}
            </div>

            <div className='min-w-0'>
              <div className='flex flex-wrap items-center gap-2'>
                <h1 className='truncate text-xl font-bold tracking-tight'>
                  {customer.name}
                </h1>
                <Badge variant='outline' className={`text-[10px] ${statusMeta.className}`}>
                  {statusMeta.label}
                </Badge>
              </div>

              <div className='mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground'>
                <button
                  type='button'
                  onClick={() =>
                    copyToClipboard(customer.customerCode || customer.id, 'Kode pelanggan')
                  }
                  className='inline-flex items-center gap-1 font-mono transition-colors hover:text-foreground'
                  title='Salin kode pelanggan'
                >
                  {customer.customerCode || customer.id}
                  <Copy className='h-3 w-3' />
                </button>
                {customer.phone && <span className='font-mono'>{customer.phone}</span>}
                {customer.email && <span className='truncate'>{customer.email}</span>}
              </div>
            </div>
          </div>

          <div className='flex flex-wrap items-center gap-2'>
            {customer.phone && (
              <Button
                variant='outline'
                size='sm'
                className='h-8 gap-1.5 border-emerald-500/30 text-xs text-emerald-600 hover:bg-emerald-500/10 hover:text-emerald-700'
                asChild
              >
                <a
                  href={`https://wa.me/${waPhone}?text=${encodeURIComponent(waMessage)}`}
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  <MessageCircle className='h-3.5 w-3.5' />
                  WhatsApp
                </a>
              </Button>
            )}

            {canManage && (
              <Button
                variant='outline'
                size='sm'
                className='h-8 gap-1.5 text-xs'
                onClick={onEdit}
              >
                <Pencil className='h-3.5 w-3.5' />
                Edit
              </Button>
            )}

            {canBilling && (
              <Button
                variant='outline'
                size='sm'
                className='h-8 gap-1.5 text-xs'
                onClick={onAddSubscription}
              >
                <Plus className='h-3.5 w-3.5' />
                Tambah Layanan
              </Button>
            )}

            {canBilling && dueInvoice && (
              <Button
                size='sm'
                className='h-8 gap-1.5 bg-emerald-600 text-xs text-white hover:bg-emerald-700'
                onClick={() => onPay(dueInvoice)}
              >
                <CreditCard className='h-3.5 w-3.5' />
                Bayar {formatCurrency(outstanding)}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
