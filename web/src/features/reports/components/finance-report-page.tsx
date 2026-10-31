import { useState } from 'react'
import { BarChart3, Landmark, Scale, Timer } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Badge } from '@/components/ui/badge'
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
import {
  useARAgingQuery,
  useGatewayReconciliationQuery,
  usePSBStatsQuery,
} from '../api/use-reports'
import { BillingOpsReports } from './billing-ops-reports'

function formatIDR(value: number | undefined) {
  return `Rp${Number(value || 0).toLocaleString('id-ID')}`
}

const AR_BUCKET_LABELS: Record<string, string> = {
  CURRENT: 'Belum jatuh tempo',
  D1_30: '1–30 hari',
  D31_60: '31–60 hari',
  D61_90: '61–90 hari',
  D90_PLUS: '> 90 hari',
}

/**
 * Laporan keuangan & operasional (planning v2 Fase 5): AR Aging, rekonsiliasi
 * transaksi gateway vs buku kas, dan statistik efisiensi PSB.
 */
export function FinanceReportPage() {
  const [asOf, setAsOf] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')

  const arQuery = useARAgingQuery(asOf)
  const reconQuery = useGatewayReconciliationQuery(from, to)
  const psbQuery = usePSBStatsQuery(from, to)

  const ar = arQuery.data
  const recon = reconQuery.data
  const psb = psbQuery.data

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div>
          <div className='flex items-center gap-2'>
            <BarChart3 className='h-6 w-6 text-primary' />
            <h1 className='text-2xl font-bold tracking-tight'>
              Laporan Keuangan & Operasional
            </h1>
          </div>
          <p className='mt-0.5 text-sm text-muted-foreground'>
            AR Aging, rekonsiliasi payment gateway terhadap buku kas, dan
            statistik efisiensi penyelesaian PSB.
          </p>
        </div>

        <div className='grid gap-4 xl:grid-cols-2'>
          {/* AR Aging */}
          <Card>
            <CardHeader className='flex-row items-center justify-between space-y-0'>
              <div>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <Scale className='h-4 w-4 text-amber-500' /> AR Aging
                </CardTitle>
                <CardDescription>
                  Umur piutang tagihan pelanggan yang masih outstanding.
                </CardDescription>
              </div>
              <div className='flex items-center gap-2'>
                <Input
                  type='date'
                  value={asOf}
                  onChange={(e) => setAsOf(e.target.value)}
                  className='h-8 w-36 text-xs'
                />
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() => {
                    setAsOf('')
                  }}
                >
                  Hari ini
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Bucket</TableHead>
                    <TableHead className='text-right'>Jumlah Tagihan</TableHead>
                    <TableHead className='text-right'>Outstanding</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(ar?.buckets || []).map((b) => (
                    <TableRow key={b.bucket}>
                      <TableCell className='text-xs'>
                        {AR_BUCKET_LABELS[b.bucket] || b.bucket}
                      </TableCell>
                      <TableCell className='text-right text-xs'>
                        {b.invoiceCount}
                      </TableCell>
                      <TableCell className='text-right text-xs font-semibold'>
                        {formatIDR(b.outstanding)}
                      </TableCell>
                    </TableRow>
                  ))}
                  {(!ar || ar.buckets.length === 0) && (
                    <TableRow>
                      <TableCell
                        colSpan={3}
                        className='py-6 text-center text-xs text-muted-foreground'
                      >
                        Tidak ada piutang outstanding.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <div className='mt-3 flex items-center justify-between rounded-md bg-muted/50 p-3 text-sm'>
                <span className='text-muted-foreground'>
                  Total piutang ({ar?.totalCount || 0} tagihan)
                </span>
                <span className='font-bold'>{formatIDR(ar?.totalOutstanding)}</span>
              </div>
            </CardContent>
          </Card>

          {/* Rekonsiliasi Gateway */}
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Landmark className='h-4 w-4 text-blue-500' /> Rekonsiliasi Gateway
              </CardTitle>
              <CardDescription>
                Perbandingan transaksi gateway settled dengan kas tercatat.
              </CardDescription>
              <div className='flex flex-wrap items-center gap-2 pt-2'>
                <Input
                  type='date'
                  value={from}
                  onChange={(e) => setFrom(e.target.value)}
                  className='h-8 w-36 text-xs'
                />
                <span className='text-xs text-muted-foreground'>s/d</span>
                <Input
                  type='date'
                  value={to}
                  onChange={(e) => setTo(e.target.value)}
                  className='h-8 w-36 text-xs'
                />
              </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Gateway</TableHead>
                    <TableHead className='text-right'>Tx</TableHead>
                    <TableHead className='text-right'>Gateway</TableHead>
                    <TableHead className='text-right'>Tercatat</TableHead>
                    <TableHead className='text-right'>Selisih</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(recon?.rows || []).map((r) => (
                    <TableRow key={r.gateway}>
                      <TableCell className='text-xs font-medium'>{r.gateway}</TableCell>
                      <TableCell className='text-right text-xs'>
                        {r.transactionCount}
                      </TableCell>
                      <TableCell className='text-right text-xs'>
                        {formatIDR(r.gatewayAmount)}
                      </TableCell>
                      <TableCell className='text-right text-xs'>
                        {formatIDR(r.recordedAmount)}
                      </TableCell>
                      <TableCell className='text-right text-xs'>
                        <Badge
                          variant='outline'
                          className={
                            Math.abs(r.difference) < 0.01
                              ? 'bg-emerald-500/15 text-emerald-600 border-emerald-500/30'
                              : 'bg-amber-500/15 text-amber-600 border-amber-500/30'
                          }
                        >
                          {formatIDR(r.difference)}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                  {(!recon || recon.rows.length === 0) && (
                    <TableRow>
                      <TableCell
                        colSpan={5}
                        className='py-6 text-center text-xs text-muted-foreground'
                      >
                        Tidak ada transaksi gateway settled pada rentang ini.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <div className='mt-3 grid grid-cols-3 gap-2 text-center text-xs'>
                <div className='rounded-md bg-muted/50 p-2'>
                  <p className='text-muted-foreground'>Total Gateway</p>
                  <p className='font-semibold'>{formatIDR(recon?.totalGateway)}</p>
                </div>
                <div className='rounded-md bg-muted/50 p-2'>
                  <p className='text-muted-foreground'>Total Tercatat</p>
                  <p className='font-semibold'>{formatIDR(recon?.totalRecorded)}</p>
                </div>
                <div className='rounded-md bg-muted/50 p-2'>
                  <p className='text-muted-foreground'>Total Fee</p>
                  <p className='font-semibold'>{formatIDR(recon?.totalFee)}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Statistik PSB */}
        <Card>
          <CardHeader>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Timer className='h-4 w-4 text-purple-500' /> Statistik Efisiensi PSB
            </CardTitle>
            <CardDescription>
              Durasi survey & instalasi teknisi serta konversi pendaftaran.
            </CardDescription>
            <div className='flex flex-wrap items-center gap-2 pt-2'>
              <Input
                type='date'
                value={from}
                onChange={(e) => setFrom(e.target.value)}
                className='h-8 w-36 text-xs'
              />
              <span className='text-xs text-muted-foreground'>s/d</span>
              <Input
                type='date'
                value={to}
                onChange={(e) => setTo(e.target.value)}
                className='h-8 w-36 text-xs'
              />
            </div>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='grid grid-cols-2 gap-3 lg:grid-cols-5'>
              {[
                { label: 'Total Pendaftaran', value: psb?.total || 0 },
                { label: 'Terkonversi', value: psb?.converted || 0 },
                { label: 'Ditutup', value: psb?.closed || 0 },
                { label: 'Rata-rata Survey', value: `${psb?.avgSurveyHours || 0} jam` },
                { label: 'Rata-rata Selesai', value: `${psb?.avgLeadHours || 0} jam` },
              ].map((stat) => (
                <div key={stat.label} className='rounded-md border p-3'>
                  <p className='text-xs text-muted-foreground'>{stat.label}</p>
                  <p className='mt-1 text-lg font-bold'>{stat.value}</p>
                </div>
              ))}
            </div>
            <div className='flex flex-wrap gap-2'>
              {(psb?.byStatus || []).map((s) => (
                <Badge key={s.status} variant='outline' className='text-xs'>
                  {s.status}: {s.count}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Laporan keuangan Billing Ops v3 */}
        <BillingOpsReports />
      </Main>
    </>
  )
}
