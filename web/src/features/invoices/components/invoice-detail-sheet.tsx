import {
  CreditCard,
  Printer,
  QrCode,
  Receipt,
  User,
  Wifi,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Card, CardContent } from '@/components/ui/card'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useCustomerQuery } from '@/features/customer/api/use-customer'
import { useSubscriptionQuery } from '@/features/billing/api/use-billing'
import { invoiceStatusBadge, ITEM_TYPE_META } from '../data/constants'
import { useInvoices } from './invoices-provider'

interface InvoiceDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  invoice: Invoice | null
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

export function InvoiceDetailSheet({
  open,
  onOpenChange,
  invoice,
}: InvoiceDetailSheetProps) {
  const { setOpen } = useInvoices()
  const customerId = invoice?.customerId || ''
  const subscriptionId = invoice?.subscriptionId || ''

  const { data: customer } = useCustomerQuery(customerId)
  const { data: subscription } = useSubscriptionQuery(subscriptionId)

  if (!invoice) return null

  const isPaid = invoice.status === 'PAID'
  const outstanding = Math.max(0, invoice.total - invoice.paidAmount)
  const statusMeta = invoiceStatusBadge(invoice.status)

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='flex w-full flex-col sm:max-w-xl md:max-w-2xl p-0'>
        <SheetHeader className='border-b p-6 pb-4'>
          <div className='flex items-start justify-between gap-4'>
            <div>
              <div className='flex items-center gap-2'>
                <SheetTitle className='text-xl font-bold font-mono'>
                  {invoice.invoiceNumber || invoice.id}
                </SheetTitle>
                <Badge variant='outline' className={`text-[11px] ${statusMeta.className}`}>
                  {statusMeta.label}
                </Badge>
              </div>
              <SheetDescription className='mt-1'>
                Faktur tagihan periode {invoice.period || '-'} · Diterbitkan {formatUnixDate(invoice.createdAtUnix)}
              </SheetDescription>
            </div>

            {/* Quick Actions */}
            <div className='flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                onClick={() => setOpen('print')}
                className='h-8 text-xs gap-1'
              >
                <Printer className='h-3.5 w-3.5' />
                Cetak
              </Button>
              {!isPaid && (
                <Button
                  size='sm'
                  onClick={() => setOpen('cashier')}
                  className='h-8 text-xs gap-1 bg-emerald-600 hover:bg-emerald-700 text-white'
                >
                  <CreditCard className='h-3.5 w-3.5' />
                  Bayar Kasir
                </Button>
              )}
            </div>
          </div>
        </SheetHeader>

        <ScrollArea className='flex-1 p-6 space-y-6'>
          {/* Info Pelanggan & Langganan */}
          <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
            {/* Pelanggan */}
            <Card className='bg-muted/20'>
              <CardContent className='p-4 space-y-1.5'>
                <div className='flex items-center gap-2 text-xs font-semibold text-foreground mb-2'>
                  <User className='h-4 w-4 text-primary' />
                  Informasi Pelanggan
                </div>
                {customer ? (
                  <>
                    <p className='text-sm font-bold text-foreground'>{customer.name}</p>
                    <p className='text-xs text-muted-foreground font-mono'>Kode: {customer.customerCode}</p>
                    <p className='text-xs text-muted-foreground'>Telp: {customer.phone || '-'}</p>
                    {customer.address && (
                      <p className='text-xs text-muted-foreground line-clamp-2'>Alamat: {customer.address}</p>
                    )}
                  </>
                ) : (
                  <p className='text-xs text-muted-foreground'>Memuat data pelanggan...</p>
                )}
              </CardContent>
            </Card>

            {/* Langganan */}
            <Card className='bg-muted/20'>
              <CardContent className='p-4 space-y-1.5'>
                <div className='flex items-center gap-2 text-xs font-semibold text-foreground mb-2'>
                  <Wifi className='h-4 w-4 text-blue-600' />
                  Layanan Internet
                </div>
                {subscription ? (
                  <>
                    <p className='text-sm font-bold text-foreground'>{subscription.planName || 'Paket Internet'}</p>
                    <p className='text-xs text-muted-foreground font-mono'>Username: {subscription.remoteUsername}</p>
                    <p className='text-xs text-muted-foreground'>Tipe: {subscription.serviceType} · Router: {subscription.deviceName || '-'}</p>
                    <p className='text-xs text-muted-foreground'>Siklus: {subscription.billingCycle || 'Bulanan'} (Tgl {subscription.billingDay})</p>
                  </>
                ) : (
                  <p className='text-xs text-muted-foreground'>Langganan tidak ditautkan atau ad-hoc.</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Rincian Item Faktur */}
          <div className='space-y-3 pt-2'>
            <h4 className='text-xs font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5'>
              <Receipt className='h-3.5 w-3.5' /> Rincian Komponen Tagihan
            </h4>

            <div className='rounded-lg border overflow-hidden bg-card'>
              <table className='w-full text-xs'>
                <thead className='bg-muted/50 border-b text-muted-foreground text-left'>
                  <tr>
                    <th className='p-3 font-medium'>Deskripsi Item</th>
                    <th className='p-3 font-medium text-center'>Qty</th>
                    <th className='p-3 font-medium text-right'>Harga Satuan</th>
                    <th className='p-3 font-medium text-right'>Jumlah</th>
                  </tr>
                </thead>
                <tbody className='divide-y'>
                  {invoice.items && invoice.items.length > 0 ? (
                    invoice.items.map((item, idx) => (
                      <tr key={item.id || idx}>
                        <td className='p-3'>
                          <p className='font-semibold text-foreground'>{item.description}</p>
                          <p className='text-[10px] text-muted-foreground'>
                            {ITEM_TYPE_META[item.itemType] || item.itemType}
                          </p>
                        </td>
                        <td className='p-3 text-center'>{item.quantity || 1}</td>
                        <td className='p-3 text-right font-mono'>{formatCurrency(item.unitPrice)}</td>
                        <td className='p-3 text-right font-mono font-semibold'>{formatCurrency(item.amount)}</td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td className='p-3'>
                        <p className='font-semibold text-foreground'>
                          Tagihan Langganan Internet Periode {invoice.period}
                        </p>
                        <p className='text-[10px] text-muted-foreground'>Paket Layanan Bulanan</p>
                      </td>
                      <td className='p-3 text-center'>1</td>
                      <td className='p-3 text-right font-mono'>{formatCurrency(invoice.subtotal || invoice.total)}</td>
                      <td className='p-3 text-right font-mono font-semibold'>{formatCurrency(invoice.subtotal || invoice.total)}</td>
                    </tr>
                  )}
                </tbody>
              </table>

              {/* Total Summary Breakdown */}
              <div className='border-t bg-muted/20 p-4 space-y-2 text-xs'>
                <div className='flex justify-between text-muted-foreground'>
                  <span>Subtotal:</span>
                  <span className='font-mono font-medium'>{formatCurrency(invoice.subtotal || invoice.total)}</span>
                </div>
                {invoice.discount > 0 && (
                  <div className='flex justify-between text-emerald-600'>
                    <span>Diskon / Potongan:</span>
                    <span className='font-mono font-medium'>- {formatCurrency(invoice.discount)}</span>
                  </div>
                )}
                {invoice.taxAmount > 0 && (
                  <div className='flex justify-between text-muted-foreground'>
                    <span>Pajak (PPN):</span>
                    <span className='font-mono font-medium'>+ {formatCurrency(invoice.taxAmount)}</span>
                  </div>
                )}
                <div className='flex justify-between text-sm font-bold border-t pt-2 text-foreground'>
                  <span>Total Tagihan:</span>
                  <span className='font-mono text-primary'>{formatCurrency(invoice.total)}</span>
                </div>
                <div className='flex justify-between text-xs text-muted-foreground'>
                  <span>Sudah Dibayar:</span>
                  <span className='font-mono font-semibold text-emerald-600'>{formatCurrency(invoice.paidAmount)}</span>
                </div>
                {!isPaid && (
                  <div className='flex justify-between text-xs font-bold text-rose-600 dark:text-rose-400 border-t pt-1'>
                    <span>Sisa Tagihan (Outstanding):</span>
                    <span className='font-mono text-sm'>{formatCurrency(outstanding)}</span>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Kode Bayar & QRIS Payload */}
          <div className='rounded-lg border bg-muted/20 p-4 space-y-3'>
            <div className='flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground'>
              <QrCode className='h-4 w-4' /> Kanal Pembayaran Pelanggan
            </div>
            <div className='grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs'>
              <div className='bg-background p-3 rounded-lg border'>
                <span className='text-muted-foreground'>Kode Bayar Manual:</span>
                <p className='font-mono text-base font-bold text-primary mt-1'>
                  {invoice.manualPaymentCode || '-'}
                </p>
                <p className='text-[10px] text-muted-foreground mt-0.5'>
                  Dapat dimasukkan pada kasir atau portal pelanggan
                </p>
              </div>

              <div className='bg-background p-3 rounded-lg border'>
                <span className='text-muted-foreground'>Jatuh Tempo Pembayaran:</span>
                <p className='font-semibold text-foreground mt-1'>
                  {formatUnixDate(invoice.dueDateUnix)}
                </p>
                {invoice.paidAtUnix ? (
                  <p className='text-[10px] text-emerald-600 mt-0.5'>
                    Lunas pada {formatUnixDate(invoice.paidAtUnix)}
                  </p>
                ) : (
                  <p className='text-[10px] text-rose-600 mt-0.5'>
                    Status: Belum Lunas
                  </p>
                )}
              </div>
            </div>

            {invoice.notes && (
              <div className='bg-background p-3 rounded-lg border text-xs'>
                <span className='text-muted-foreground font-semibold'>Catatan Tagihan:</span>
                <p className='mt-1 text-foreground'>{invoice.notes}</p>
              </div>
            )}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
