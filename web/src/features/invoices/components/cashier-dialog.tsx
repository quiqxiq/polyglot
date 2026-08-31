import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import {
  CheckCircle2,
  CreditCard,
  Globe,
  Printer,
  QrCode,
  Search,
  User,
  Wallet,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent } from '@/components/ui/card'
import { type Invoice, CashierPayRequest, ResolveMethod } from '@/gen/v1/billing_pb'
import { useCashAccountsQuery, useCashCategoriesQuery } from '@/features/cashbook/api/use-cashbook'
import { useCustomerQuery } from '@/features/customer/api/use-customer'
import { useCashierPayMutation, useCashierResolveQuery } from '../api/use-invoices'
import { cashierPaySchema, type CashierPayFormValues } from '../data/schema'
import { invoiceStatusBadge } from '../data/constants'
import { useInvoices } from './invoices-provider'

interface CashierDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentInvoice?: Invoice | null
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

export function CashierDialog({
  open,
  onOpenChange,
  currentInvoice,
}: CashierDialogProps) {
  const { setOpen, setCurrentInvoice } = useInvoices()
  const [resolveMethod, setResolveMethod] = useState<ResolveMethod>(ResolveMethod.RESOLVE_CODE)
  const [identifierInput, setIdentifierInput] = useState('')
  const [resolvedInvoice, setResolvedInvoice] = useState<Invoice | null>(currentInvoice || null)
  const [paymentSuccess, setPaymentSuccess] = useState<{ paymentNo: string; invoice: Invoice } | null>(null)

  const { data: accounts = [] } = useCashAccountsQuery(true)
  const { data: categories = [] } = useCashCategoriesQuery(true)
  const cashierPay = useCashierPayMutation()

  // Ambil data customer dari invoice yang dipilih
  const customerId = resolvedInvoice?.customerId || ''
  const { data: customer } = useCustomerQuery(customerId)

  // Query pencarian kasir
  const {
    data: resolveResult,
    isFetching: isResolving,
    refetch: triggerResolve,
  } = useCashierResolveQuery(identifierInput.trim(), resolveMethod, false)

  const form = useForm<CashierPayFormValues>({
    resolver: zodResolver(cashierPaySchema),
    defaultValues: {
      invoiceId: resolvedInvoice?.id || '',
      amount: resolvedInvoice ? Math.max(0, resolvedInvoice.total - resolvedInvoice.paidAmount) : 0,
      cashAccountId: accounts[0]?.id || '',
      incomeCategoryId: categories.find((c) => c.type === 'INCOME')?.id || '',
      scanMethod: 'CODE_INPUT',
      reference: '',
      notes: '',
    },
  })

  // Sinkronisasi saat invoice dipilih dari tabel
  useEffect(() => {
    if (currentInvoice) {
      setResolvedInvoice(currentInvoice)
      setIdentifierInput(currentInvoice.manualPaymentCode || currentInvoice.invoiceNumber)
      const outstanding = Math.max(0, currentInvoice.total - currentInvoice.paidAmount)
      form.setValue('invoiceId', currentInvoice.id)
      form.setValue('amount', outstanding)
    }
  }, [currentInvoice, form])

  // Set default akun & kategori saat data dimuat
  useEffect(() => {
    if (accounts.length > 0 && !form.getValues('cashAccountId')) {
      form.setValue('cashAccountId', accounts[0].id)
    }
    const incCat = categories.find((c) => c.type === 'INCOME')
    if (incCat && !form.getValues('incomeCategoryId')) {
      form.setValue('incomeCategoryId', incCat.id)
    }
  }, [accounts, categories, form])

  // Tangani hasil resolve manual
  const handleResolveSearch = async (e?: React.FormEvent) => {
    if (e) e.preventDefault()
    if (!identifierInput.trim()) {
      toast.error('Masukkan kode pembayaran, payload QR, atau kode portal')
      return
    }

    try {
      const res = await triggerResolve()
      if (res.data?.invoice) {
        setResolvedInvoice(res.data.invoice)
        const outstanding = Math.max(0, res.data.invoice.total - res.data.invoice.paidAmount)
        form.setValue('invoiceId', res.data.invoice.id)
        form.setValue('amount', outstanding)
        toast.success(`Faktur ${res.data.invoice.invoiceNumber} ditemukan!`)
      } else {
        toast.error('Tagihan tidak ditemukan atau sudah lunas')
      }
    } catch (err: any) {
      toast.error('Gagal mencari tagihan', {
        description: err?.message || 'Kode tidak ditemukan pada sistem',
      })
    }
  }

  const onSubmit = async (values: CashierPayFormValues) => {
    if (!resolvedInvoice) {
      toast.error('Pilih atau cari faktur tagihan terlebih dahulu')
      return
    }

    try {
      const res = await cashierPay.mutateAsync(
        new CashierPayRequest({
          invoiceId: values.invoiceId,
          amount: Number(values.amount),
          cashAccountId: values.cashAccountId,
          incomeCategoryId: values.incomeCategoryId,
          scanMethod: values.scanMethod,
          reference: values.reference,
          notes: values.notes,
        })
      )

      const paidInv = res.invoice || resolvedInvoice
      setPaymentSuccess({
        paymentNo: res.paymentNo || 'PAY-SUCCESS',
        invoice: paidInv,
      })
      toast.success(`Pembayaran berhasil! No. Kwitansi: ${res.paymentNo}`)
    } catch (err: any) {
      toast.error('Pembayaran gagal diproses', {
        description: err?.message || 'Terjadi kesalahan transaksi di database',
      })
    }
  }

  const handleReset = () => {
    setPaymentSuccess(null)
    setResolvedInvoice(null)
    setIdentifierInput('')
    form.reset()
  }

  const outstanding = resolvedInvoice
    ? Math.max(0, resolvedInvoice.total - resolvedInvoice.paidAmount)
    : 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-xl max-h-[90vh] overflow-y-auto'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <CreditCard className='h-5 w-5 text-emerald-600' />
            <DialogTitle>Kasir POS — Pembayaran Cepat Tagihan</DialogTitle>
          </div>
          <DialogDescription>
            Penerimaan pembayaran kasir, pencatatan otomatis jurnal kas, kuitansi WA, dan pemulihan isolir router.
          </DialogDescription>
        </DialogHeader>

        {paymentSuccess ? (
          /* ─── State Sukses Pembayaran ─── */
          <div className='py-6 text-center space-y-4'>
            <div className='mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600'>
              <CheckCircle2 className='h-10 w-10' />
            </div>
            <div>
              <h3 className='text-lg font-bold text-foreground'>Pembayaran Berhasil Diterima</h3>
              <p className='font-mono text-sm font-semibold text-emerald-600 dark:text-emerald-400 mt-1'>
                No. Kwitansi: {paymentSuccess.paymentNo}
              </p>
              <p className='text-xs text-muted-foreground mt-1'>
                Jurnal kas masuk telah dicatat otomatis & notifikasi WhatsApp diteruskan ke antrean pelanggan.
              </p>
            </div>

            <Card className='bg-muted/30 border text-left'>
              <CardContent className='p-4 space-y-2 text-xs'>
                <div className='flex justify-between'>
                  <span className='text-muted-foreground'>Nomor Faktur:</span>
                  <span className='font-mono font-bold'>{paymentSuccess.invoice.invoiceNumber}</span>
                </div>
                <div className='flex justify-between'>
                  <span className='text-muted-foreground'>Total Pembayaran:</span>
                  <span className='font-mono font-bold text-emerald-600'>
                    {formatCurrency(form.getValues('amount'))}
                  </span>
                </div>
                <div className='flex justify-between'>
                  <span className='text-muted-foreground'>Status Faktur:</span>
                  <Badge variant='outline' className='bg-emerald-500/15 text-emerald-700 border-emerald-500/30 text-[10px]'>
                    {paymentSuccess.invoice.status}
                  </Badge>
                </div>
              </CardContent>
            </Card>

            <div className='flex justify-center gap-3 pt-2'>
              <Button
                variant='outline'
                onClick={() => {
                  setCurrentInvoice(paymentSuccess.invoice)
                  setOpen('print')
                }}
                className='gap-1.5'
              >
                <Printer className='h-4 w-4' />
                Cetak Kwitansi
              </Button>
              <Button onClick={handleReset} className='bg-emerald-600 hover:bg-emerald-700 text-white'>
                Transaksi Baru
              </Button>
            </div>
          </div>
        ) : (
          /* ─── State Form Kasir ─── */
          <div className='space-y-4 py-1'>
            {/* 1. Bar Pencarian / Lookup */}
            {!currentInvoice && (
              <div className='space-y-2'>
                <label className='text-xs font-medium text-muted-foreground'>Cari Tagihan Pelanggan</label>
                <Tabs
                  value={String(resolveMethod)}
                  onValueChange={(val) => setResolveMethod(Number(val) as ResolveMethod)}
                >
                  <TabsList className='grid w-full grid-cols-3 h-8'>
                    <TabsTrigger value={String(ResolveMethod.RESOLVE_CODE)} className='text-xs gap-1'>
                      <CreditCard className='h-3 w-3' /> Kode Bayar
                    </TabsTrigger>
                    <TabsTrigger value={String(ResolveMethod.RESOLVE_QR)} className='text-xs gap-1'>
                      <QrCode className='h-3 w-3' /> QRIS Payload
                    </TabsTrigger>
                    <TabsTrigger value={String(ResolveMethod.RESOLVE_PORTAL)} className='text-xs gap-1'>
                      <Globe className='h-3 w-3' /> Kode Portal
                    </TabsTrigger>
                  </TabsList>
                </Tabs>

                <form onSubmit={handleResolveSearch} className='flex gap-2 pt-1'>
                  <Input
                    placeholder={
                      resolveMethod === ResolveMethod.RESOLVE_CODE
                        ? 'Masukkan kode bayar manual / No Faktur...'
                        : resolveMethod === ResolveMethod.RESOLVE_QR
                        ? 'Scan atau tempel payload QRIS...'
                        : 'Masukkan kode akses portal pelanggan...'
                    }
                    className='h-9 text-xs font-mono'
                    value={identifierInput}
                    onChange={(e) => setIdentifierInput(e.target.value)}
                  />
                  <Button type='submit' size='sm' className='h-9 gap-1 shrink-0' disabled={isResolving}>
                    <Search className='h-3.5 w-3.5' />
                    {isResolving ? 'Mencari...' : 'Cari'}
                  </Button>
                </form>
              </div>
            )}

            {/* 2. Kartu Rincian Tagihan Terpilih */}
            {resolvedInvoice && (
              <Card className='border-l-4 border-l-emerald-500 bg-muted/20'>
                <CardContent className='p-4 space-y-2.5'>
                  <div className='flex items-start justify-between'>
                    <div>
                      <div className='flex items-center gap-2'>
                        <span className='font-mono text-sm font-bold text-foreground'>
                          {resolvedInvoice.invoiceNumber || resolvedInvoice.id}
                        </span>
                        <Badge
                          variant='outline'
                          className={`text-[10px] ${invoiceStatusBadge(resolvedInvoice.status).className}`}
                        >
                          {invoiceStatusBadge(resolvedInvoice.status).label}
                        </Badge>
                      </div>
                      <p className='text-xs text-muted-foreground mt-0.5'>
                        Periode Tagihan: <span className='font-semibold'>{resolvedInvoice.period}</span> · Jatuh tempo {formatUnixDate(resolvedInvoice.dueDateUnix)}
                      </p>
                    </div>

                    <div className='text-right'>
                      <p className='text-[11px] text-muted-foreground'>Total Tagihan</p>
                      <p className='font-mono text-sm font-bold'>{formatCurrency(resolvedInvoice.total)}</p>
                    </div>
                  </div>

                  {/* Info Pelanggan */}
                  {(customer || resolveResult?.customerName) && (
                    <div className='flex items-center gap-2 rounded-lg bg-background p-2.5 border text-xs'>
                      <User className='h-4 w-4 text-primary shrink-0' />
                      <div className='min-w-0 flex-1'>
                        <p className='font-semibold text-foreground truncate'>
                          {customer?.name || resolveResult?.customerName}
                        </p>
                        <p className='text-[10px] text-muted-foreground truncate'>
                          {customer?.phone || resolveResult?.customerPhone} · {customer?.customerCode || 'Pelanggan'}
                        </p>
                      </div>
                    </div>
                  )}

                  {/* Sisa Piutang */}
                  <div className='flex items-center justify-between border-t pt-2 text-xs'>
                    <span className='text-muted-foreground'>Sisa yang harus dibayar:</span>
                    <span className='font-mono text-sm font-bold text-rose-600 dark:text-rose-400'>
                      {formatCurrency(outstanding)}
                    </span>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* 3. Form Eksekusi Pembayaran */}
            {resolvedInvoice && (
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-3.5'>
                  <div className='grid grid-cols-1 sm:grid-cols-2 gap-3'>
                    {/* Nominal Bayar */}
                    <FormField
                      control={form.control}
                      name='amount'
                      render={({ field }) => (
                        <FormItem className='sm:col-span-2'>
                          <FormLabel>Nominal Pembayaran (Rp)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              min={1}
                              max={outstanding}
                              step='any'
                              className='font-mono text-sm font-bold text-emerald-600 dark:text-emerald-400'
                              {...field}
                              value={field.value || ''}
                              onChange={(e) => field.onChange(Number(e.target.value))}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Rekening Kas Penampung */}
                    <FormField
                      control={form.control}
                      name='cashAccountId'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel className='flex items-center gap-1'>
                            <Wallet className='h-3.5 w-3.5 text-emerald-600' />
                            Rekening Kas Penampung
                          </FormLabel>
                          <Select onValueChange={field.onChange} value={field.value}>
                            <FormControl>
                              <SelectTrigger className='text-xs'>
                                <SelectValue placeholder='Pilih rekening kas' />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              {accounts.map((acc) => (
                                <SelectItem key={acc.id} value={acc.id}>
                                  {acc.name} ({acc.type === 'BANK' ? 'Bank' : 'Kasir'})
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Kategori Kas Pendapatan */}
                    <FormField
                      control={form.control}
                      name='incomeCategoryId'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Kategori Pos Pendapatan</FormLabel>
                          <Select onValueChange={field.onChange} value={field.value}>
                            <FormControl>
                              <SelectTrigger className='text-xs'>
                                <SelectValue placeholder='Pilih pos pendapatan' />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              {categories
                                .filter((c) => c.type === 'INCOME')
                                .map((cat) => (
                                  <SelectItem key={cat.id} value={cat.id}>
                                    {cat.name}
                                  </SelectItem>
                                ))}
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Metode Scan / Input */}
                    <FormField
                      control={form.control}
                      name='scanMethod'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Metode Transaksi</FormLabel>
                          <Select onValueChange={field.onChange} value={field.value}>
                            <FormControl>
                              <SelectTrigger className='text-xs'>
                                <SelectValue placeholder='Pilih metode' />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value='CODE_INPUT'>Kode Bayar / Input Manual</SelectItem>
                              <SelectItem value='QR_SCAN'>QRIS / Scan QR Code</SelectItem>
                              <SelectItem value='MANUAL'>Kasir Tunai di Kantor</SelectItem>
                              <SelectItem value='PAYMENT_GATEWAY'>Transfer Bank / Gateway</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    {/* Nomor Referensi */}
                    <FormField
                      control={form.control}
                      name='reference'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>No. Referensi / Bukti (Opsional)</FormLabel>
                          <FormControl>
                            <Input placeholder='Contoh: REF-BCA-98124' className='text-xs font-mono' {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  {/* Catatan */}
                  <FormField
                    control={form.control}
                    name='notes'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Catatan Tambahan (Opsional)</FormLabel>
                        <FormControl>
                          <Textarea placeholder='Keterangan kasir...' rows={2} className='text-xs' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <DialogFooter className='pt-3'>
                    <Button
                      type='button'
                      variant='outline'
                      onClick={() => onOpenChange(false)}
                      disabled={cashierPay.isPending}
                    >
                      Batal
                    </Button>
                    <Button
                      type='submit'
                      disabled={cashierPay.isPending || form.watch('amount') <= 0}
                      className='gap-1.5'
                    >
                      <CreditCard className='h-4 w-4' />
                      {cashierPay.isPending ? 'Memproses Transaksi...' : `Terima Pembayaran ${formatCurrency(form.watch('amount'))}`}
                    </Button>
                  </DialogFooter>
                </form>
              </Form>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
