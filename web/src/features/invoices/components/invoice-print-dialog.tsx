import { useRef, useState } from 'react'
import { Printer, Receipt, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useCustomerQuery } from '@/features/customer/api/use-customer'

interface InvoicePrintDialogProps {
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
    month: 'short',
    year: 'numeric',
  })
}

export function InvoicePrintDialog({
  open,
  onOpenChange,
  invoice,
}: InvoicePrintDialogProps) {
  const [printFormat, setPrintFormat] = useState<'A4' | 'THERMAL'>('A4')
  const printRef = useRef<HTMLDivElement>(null)

  const customerId = invoice?.customerId || ''
  const { data: customer } = useCustomerQuery(customerId)

  if (!invoice) return null

  const isPaid = invoice.status === 'PAID'
  const outstanding = Math.max(0, invoice.total - invoice.paidAmount)

  const handlePrint = () => {
    window.print()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-2xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <Printer className='h-5 w-5 text-primary' />
              <DialogTitle>
                {isPaid ? 'Cetak Kwitansi Pembayaran' : 'Cetak Faktur Tagihan'}
              </DialogTitle>
            </div>
            {/* Format Selector */}
            <Tabs
              value={printFormat}
              onValueChange={(val) => setPrintFormat(val as 'A4' | 'THERMAL')}
              className='mr-6'
            >
              <TabsList className='h-8'>
                <TabsTrigger value='A4' className='text-xs gap-1'>
                  <FileText className='h-3 w-3' /> Format A4
                </TabsTrigger>
                <TabsTrigger value='THERMAL' className='text-xs gap-1'>
                  <Receipt className='h-3 w-3' /> Struk Termal
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
          <DialogDescription>
            Pratinjau cetak dokumen tagihan atau bukti pembayaran kasir.
          </DialogDescription>
        </DialogHeader>

        {/* ─── Printable Area ─── */}
        <div ref={printRef} className='py-2'>
          {printFormat === 'A4' ? (
            /* ─── Layout A4 Standard ─── */
            <div className='rounded-xl border bg-card p-6 text-xs space-y-6 shadow-xs font-sans text-foreground'>
              {/* Header ISP & Faktur */}
              <div className='flex justify-between border-b pb-4'>
                <div>
                  <h3 className='text-lg font-bold tracking-tight text-primary'>POLYGLOT NETOPS</h3>
                  <p className='text-muted-foreground'>Internet Service & Network Operations</p>
                  <p className='text-muted-foreground text-[11px] mt-1'>Customer Care: 0812-3456-7890</p>
                </div>
                <div className='text-right'>
                  <h4 className='text-base font-bold font-mono'>
                    {isPaid ? 'KWITANSI PEMBAYARAN' : 'FAKTUR TAGIHAN'}
                  </h4>
                  <p className='font-mono font-semibold text-primary mt-0.5'>
                    {invoice.invoiceNumber || invoice.id}
                  </p>
                  <p className='text-muted-foreground text-[11px] mt-0.5'>
                    Periode: <span className='font-semibold text-foreground'>{invoice.period}</span>
                  </p>
                </div>
              </div>

              {/* Ditagihkan Kepada */}
              <div className='grid grid-cols-2 gap-4 border-b pb-4'>
                <div>
                  <p className='font-bold uppercase tracking-wider text-muted-foreground text-[10px]'>
                    DITAGIHKAN KEPADA:
                  </p>
                  <p className='font-bold text-sm text-foreground mt-1'>{customer?.name || 'Pelanggan'}</p>
                  <p className='text-muted-foreground font-mono'>Kode: {customer?.customerCode || '-'}</p>
                  <p className='text-muted-foreground'>Telp: {customer?.phone || '-'}</p>
                  {customer?.address && <p className='text-muted-foreground mt-0.5'>{customer.address}</p>}
                </div>
                <div className='text-right space-y-1'>
                  <p className='font-bold uppercase tracking-wider text-muted-foreground text-[10px]'>
                    STATUS & TANGGAL:
                  </p>
                  <p>
                    Tanggal Terbit:{' '}
                    <span className='font-semibold'>{formatUnixDate(invoice.createdAtUnix)}</span>
                  </p>
                  <p>
                    Jatuh Tempo:{' '}
                    <span className='font-semibold'>{formatUnixDate(invoice.dueDateUnix)}</span>
                  </p>
                  <p>
                    Status:{' '}
                    <span className={`font-bold ${isPaid ? 'text-emerald-600' : 'text-rose-600'}`}>
                      {isPaid ? 'LUNAS' : invoice.status}
                    </span>
                  </p>
                  {invoice.paidAtUnix ? (
                    <p className='text-emerald-600 font-semibold'>
                      Dibayar: {formatUnixDate(invoice.paidAtUnix)}
                    </p>
                  ) : null}
                </div>
              </div>

              {/* Rincian Tabel */}
              <table className='w-full'>
                <thead>
                  <tr className='border-b bg-muted/30 text-muted-foreground font-semibold text-left'>
                    <th className='py-2 px-3'>Deskripsi Layanan</th>
                    <th className='py-2 px-3 text-center'>Qty</th>
                    <th className='py-2 px-3 text-right'>Harga Satuan</th>
                    <th className='py-2 px-3 text-right'>Jumlah</th>
                  </tr>
                </thead>
                <tbody className='divide-y'>
                  {invoice.items && invoice.items.length > 0 ? (
                    invoice.items.map((item, i) => (
                      <tr key={i}>
                        <td className='py-2.5 px-3 font-medium'>{item.description}</td>
                        <td className='py-2.5 px-3 text-center'>{item.quantity || 1}</td>
                        <td className='py-2.5 px-3 text-right font-mono'>{formatCurrency(item.unitPrice)}</td>
                        <td className='py-2.5 px-3 text-right font-mono font-semibold'>{formatCurrency(item.amount)}</td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td className='py-2.5 px-3 font-medium'>Tagihan Layanan Internet Periode {invoice.period}</td>
                      <td className='py-2.5 px-3 text-center'>1</td>
                      <td className='py-2.5 px-3 text-right font-mono'>{formatCurrency(invoice.subtotal || invoice.total)}</td>
                      <td className='py-2.5 px-3 text-right font-mono font-semibold'>{formatCurrency(invoice.subtotal || invoice.total)}</td>
                    </tr>
                  )}
                </tbody>
              </table>

              {/* Ringkasan Total */}
              <div className='flex justify-end border-t pt-4'>
                <div className='w-64 space-y-1.5'>
                  <div className='flex justify-between text-muted-foreground'>
                    <span>Subtotal:</span>
                    <span className='font-mono font-medium'>{formatCurrency(invoice.subtotal || invoice.total)}</span>
                  </div>
                  {invoice.discount > 0 && (
                    <div className='flex justify-between text-emerald-600'>
                      <span>Diskon:</span>
                      <span className='font-mono'>- {formatCurrency(invoice.discount)}</span>
                    </div>
                  )}
                  {invoice.taxAmount > 0 && (
                    <div className='flex justify-between text-muted-foreground'>
                      <span>PPN:</span>
                      <span className='font-mono'>+ {formatCurrency(invoice.taxAmount)}</span>
                    </div>
                  )}
                  <div className='flex justify-between text-sm font-bold border-t pt-2'>
                    <span>Total:</span>
                    <span className='font-mono text-primary'>{formatCurrency(invoice.total)}</span>
                  </div>
                  <div className='flex justify-between text-muted-foreground'>
                    <span>Dibayar:</span>
                    <span className='font-mono text-emerald-600 font-semibold'>{formatCurrency(invoice.paidAmount)}</span>
                  </div>
                  {!isPaid && (
                    <div className='flex justify-between font-bold text-rose-600 border-t pt-1'>
                      <span>Sisa:</span>
                      <span className='font-mono'>{formatCurrency(outstanding)}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Footer Note */}
              <div className='border-t pt-4 text-center text-muted-foreground text-[10px] space-y-0.5'>
                <p>Terima kasih atas kepercayaan Anda menggunakan layanan internet kami.</p>
                <p>Dokumen ini diterbitkan secara elektronik oleh Polyglot NetOps Engine dan sah tanpa tanda tangan basah.</p>
              </div>
            </div>
          ) : (
            /* ─── Layout Struk Termal 58mm/80mm ─── */
            <div className='mx-auto max-w-[320px] rounded-lg border bg-card p-4 font-mono text-[11px] space-y-3 shadow-xs'>
              <div className='text-center border-b pb-2 space-y-0.5'>
                <p className='font-bold text-sm'>POLYGLOT NETOPS</p>
                <p className='text-[10px] text-muted-foreground'>Bukti Pembayaran Internet</p>
                <p className='text-[10px] text-muted-foreground'>CS: 0812-3456-7890</p>
              </div>

              <div className='space-y-1 text-[10px] border-b pb-2'>
                <div className='flex justify-between'>
                  <span>Faktur:</span>
                  <span className='font-bold'>{invoice.invoiceNumber}</span>
                </div>
                <div className='flex justify-between'>
                  <span>Pelanggan:</span>
                  <span className='font-semibold'>{customer?.name || 'Pelanggan'}</span>
                </div>
                <div className='flex justify-between'>
                  <span>Kode Cust:</span>
                  <span>{customer?.customerCode || '-'}</span>
                </div>
                <div className='flex justify-between'>
                  <span>Periode:</span>
                  <span>{invoice.period}</span>
                </div>
                <div className='flex justify-between'>
                  <span>Waktu:</span>
                  <span>{formatUnixDate(invoice.paidAtUnix || invoice.createdAtUnix)}</span>
                </div>
              </div>

              <div className='space-y-1 border-b pb-2 text-[10px]'>
                <div className='flex justify-between font-bold'>
                  <span>Layanan Internet</span>
                  <span>{formatCurrency(invoice.total)}</span>
                </div>
                <div className='flex justify-between text-muted-foreground'>
                  <span>Status:</span>
                  <span className={isPaid ? 'text-emerald-600 font-bold' : 'text-rose-600'}>
                    {isPaid ? 'LUNAS' : invoice.status}
                  </span>
                </div>
              </div>

              <div className='text-center text-[9px] text-muted-foreground pt-1 space-y-0.5'>
                <p>SIMPAN STRUK INI</p>
                <p>SEBAGAI BUKTI PEMBAYARAN SAH</p>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className='pt-2'>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Tutup
          </Button>
          <Button onClick={handlePrint} className='gap-1.5'>
            <Printer className='h-4 w-4' />
            Cetak Sekarang
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
