import { useState } from 'react'
import { Download, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { downloadBlob } from '@/features/billing/api/use-billing-ops'
import {
  useAgingDetailQuery,
  useCashflowReportQuery,
  useClosingReportQuery,
  useCollectionReportQuery,
  useExportReportMutation,
  useSalesReportQuery,
  useTaxReportQuery,
} from '../api/use-reports'

function formatIDR(value: number | undefined) {
  return `Rp${Number(value || 0).toLocaleString('id-ID')}`
}

function currentPeriod() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function ExportButton({
  reportType,
  params,
}: {
  reportType: string
  params: Record<string, string>
}) {
  const exportReport = useExportReportMutation()
  return (
    <Button
      variant='outline'
      size='sm'
      className='gap-1.5'
      disabled={exportReport.isPending}
      onClick={async () => {
        try {
          const res = await exportReport.mutateAsync({
            reportType,
            format: 'xlsx',
            params,
          })
          downloadBlob(res.payload, res.contentType, res.filename)
        } catch (err) {
          toast.error(`Gagal ekspor: ${(err as Error).message}`)
        }
      }}
    >
      <Download className='h-4 w-4' /> Ekspor XLSX
    </Button>
  )
}

function Loading() {
  return (
    <div className='flex items-center gap-2 p-4 text-sm text-muted-foreground'>
      <Loader2 className='h-4 w-4 animate-spin' /> Memuat laporan…
    </div>
  )
}

export function BillingOpsReports() {
  const [period, setPeriod] = useState(currentPeriod())
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [asOf, setAsOf] = useState('')

  const tax = useTaxReportQuery(period)
  const sales = useSalesReportQuery(period)
  const collection = useCollectionReportQuery(period)
  const cashflow = useCashflowReportQuery(from, to)
  const aging = useAgingDetailQuery(asOf, 200)
  const closing = useClosingReportQuery(period)

  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>Laporan Keuangan Operasional</CardTitle>
        <CardDescription>
          Pajak PPN, penjualan per paket, rasio penagihan, arus kas, rincian
          piutang, dan ringkasan tutup buku.
        </CardDescription>
        <div className='flex flex-wrap items-center gap-2 pt-2'>
          <Input
            className='h-8 w-36 text-xs'
            placeholder='Periode YYYY-MM'
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
          />
          <Input
            type='date'
            className='h-8 w-36 text-xs'
            value={from}
            onChange={(e) => setFrom(e.target.value)}
          />
          <span className='text-xs text-muted-foreground'>s/d</span>
          <Input
            type='date'
            className='h-8 w-36 text-xs'
            value={to}
            onChange={(e) => setTo(e.target.value)}
          />
          <Input
            type='date'
            className='h-8 w-36 text-xs'
            value={asOf}
            onChange={(e) => setAsOf(e.target.value)}
          />
        </div>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue='tax'>
          <TabsList className='flex h-auto flex-wrap justify-start'>
            <TabsTrigger value='tax'>Pajak PPN</TabsTrigger>
            <TabsTrigger value='sales'>Penjualan</TabsTrigger>
            <TabsTrigger value='collection'>Penagihan</TabsTrigger>
            <TabsTrigger value='cashflow'>Arus Kas</TabsTrigger>
            <TabsTrigger value='aging'>Rincian Piutang</TabsTrigger>
            <TabsTrigger value='closing'>Tutup Buku</TabsTrigger>
          </TabsList>

          <TabsContent value='tax' className='space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='grid flex-1 grid-cols-2 gap-3 md:grid-cols-4'>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Subtotal</p>
                  <p className='font-semibold'>{formatIDR(tax.data?.totalSubtotal)}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Total PPN</p>
                  <p className='font-semibold'>{formatIDR(tax.data?.totalTax)}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Total Tagihan</p>
                  <p className='font-semibold'>{formatIDR(tax.data?.totalBilled)}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>PPN Terbayar</p>
                  <p className='font-semibold'>{formatIDR(tax.data?.totalPaidTax)}</p>
                </div>
              </div>
              <ExportButton reportType='TAX' params={{ period }} />
            </div>
            {tax.isLoading ? (
              <Loading />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Periode</TableHead>
                    <TableHead className='text-right'>Faktur</TableHead>
                    <TableHead className='text-right'>Subtotal</TableHead>
                    <TableHead className='text-right'>Tarif</TableHead>
                    <TableHead className='text-right'>PPN</TableHead>
                    <TableHead className='text-right'>Total</TableHead>
                    <TableHead className='text-right'>PPN Terbayar</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(tax.data?.rows || []).map((r) => (
                    <TableRow key={r.period}>
                      <TableCell>{r.period}</TableCell>
                      <TableCell className='text-right'>{r.invoiceCount}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.subtotal)}</TableCell>
                      <TableCell className='text-right'>{r.taxPercent}%</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.taxAmount)}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.total)}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.paidTax)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>

          <TabsContent value='sales' className='space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='grid grid-cols-2 gap-3 md:grid-cols-4'>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Total Tagihan</p>
                  <p className='font-semibold'>{formatIDR(sales.data?.totalBilled)}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Total Terbayar</p>
                  <p className='font-semibold'>{formatIDR(sales.data?.totalPaid)}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Faktur Lunas</p>
                  <p className='font-semibold'>{sales.data?.paidCount || 0}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Pelanggan</p>
                  <p className='font-semibold'>{sales.data?.customerCount || 0}</p>
                </div>
              </div>
              <ExportButton reportType='SALES' params={{ period }} />
            </div>
            {sales.isLoading ? (
              <Loading />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Paket</TableHead>
                    <TableHead>Tipe</TableHead>
                    <TableHead className='text-right'>Faktur</TableHead>
                    <TableHead className='text-right'>Tagihan</TableHead>
                    <TableHead className='text-right'>Terbayar</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(sales.data?.rows || []).map((r) => (
                    <TableRow key={`${r.planName}-${r.serviceType}`}>
                      <TableCell>{r.planName}</TableCell>
                      <TableCell>{r.serviceType}</TableCell>
                      <TableCell className='text-right'>{r.invoiceCount}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.totalBilled)}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.totalPaid)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>

          <TabsContent value='collection' className='space-y-3'>
            <div className='grid grid-cols-2 gap-3 md:grid-cols-4'>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Collection Rate</p>
                <p className='font-semibold'>
                  {(collection.data?.collectionRate || 0).toFixed(1)}%
                </p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Faktur Lunas</p>
                <p className='font-semibold'>
                  {collection.data?.paidInvoices || 0}/{collection.data?.totalInvoices || 0}
                </p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Lewat Jatuh Tempo</p>
                <p className='font-semibold text-destructive'>
                  {collection.data?.overdueInvoices || 0}
                </p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Outstanding</p>
                <p className='font-semibold'>{formatIDR(collection.data?.totalOutstanding)}</p>
              </div>
            </div>
            <p className='text-xs text-muted-foreground'>
              Keterlambatan ditangani lewat pengingat, laporan umur piutang, dan
              isolir — bukan denda.
            </p>
          </TabsContent>

          <TabsContent value='cashflow' className='space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='grid grid-cols-3 gap-3'>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Kas Masuk</p>
                  <p className='font-semibold text-emerald-600'>
                    {formatIDR(cashflow.data?.totalIn)}
                  </p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Kas Keluar</p>
                  <p className='font-semibold text-destructive'>
                    {formatIDR(cashflow.data?.totalOut)}
                  </p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Net</p>
                  <p className='font-semibold'>{formatIDR(cashflow.data?.netCashflow)}</p>
                </div>
              </div>
              <ExportButton reportType='CASHFLOW' params={{ from, to }} />
            </div>
            {cashflow.isLoading ? (
              <Loading />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Kategori</TableHead>
                    <TableHead>Arah</TableHead>
                    <TableHead className='text-right'>Transaksi</TableHead>
                    <TableHead className='text-right'>Nominal</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(cashflow.data?.rows || []).map((r, i) => (
                    <TableRow key={`${r.categoryName}-${r.direction}-${i}`}>
                      <TableCell>{r.categoryName}</TableCell>
                      <TableCell>{r.direction === 'IN' ? 'Masuk' : 'Keluar'}</TableCell>
                      <TableCell className='text-right'>{r.trxCount}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.totalAmount)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>

          <TabsContent value='aging' className='space-y-3'>
            <div className='flex items-center justify-between'>
              <div className='grid grid-cols-2 gap-3'>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Pelanggan Menunggak</p>
                  <p className='font-semibold'>{aging.data?.totalCustomers || 0}</p>
                </div>
                <div className='rounded-md border p-2'>
                  <p className='text-xs text-muted-foreground'>Total Piutang</p>
                  <p className='font-semibold'>{formatIDR(aging.data?.totalOutstanding)}</p>
                </div>
              </div>
              <ExportButton reportType='AGING_DETAIL' params={{ as_of: asOf }} />
            </div>
            {aging.isLoading ? (
              <Loading />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Pelanggan</TableHead>
                    <TableHead>Telepon</TableHead>
                    <TableHead className='text-right'>Faktur</TableHead>
                    <TableHead className='text-right'>Piutang</TableHead>
                    <TableHead className='text-right'>Hari</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(aging.data?.rows || []).map((r) => (
                    <TableRow key={r.customerId}>
                      <TableCell>{r.customerName}</TableCell>
                      <TableCell>{r.phone || '-'}</TableCell>
                      <TableCell className='text-right'>{r.invoiceCount}</TableCell>
                      <TableCell className='text-right'>{formatIDR(r.outstanding)}</TableCell>
                      <TableCell className='text-right'>{r.daysOverdue}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </TabsContent>

          <TabsContent value='closing' className='space-y-3'>
            <div className='grid grid-cols-2 gap-3 md:grid-cols-4'>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Tagihan Periode</p>
                <p className='font-semibold'>{formatIDR(closing.data?.totalBilled)}</p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Tertagih</p>
                <p className='font-semibold'>{formatIDR(closing.data?.totalCollected)}</p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Outstanding</p>
                <p className='font-semibold'>{formatIDR(closing.data?.totalOutstanding)}</p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Pelanggan Aktif</p>
                <p className='font-semibold'>{closing.data?.activeCustomers || 0}</p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Langganan Baru</p>
                <p className='font-semibold'>{closing.data?.newSubscriptions || 0}</p>
              </div>
              <div className='rounded-md border p-2'>
                <p className='text-xs text-muted-foreground'>Hutang Outstanding</p>
                <p className='font-semibold'>{formatIDR(closing.data?.debtsOutstanding)}</p>
              </div>
            </div>
            <p className='text-xs text-muted-foreground'>
              Tanggal tutup buku mengikuti periode {period} (akhir bulan).
            </p>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
