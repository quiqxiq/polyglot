import { useMemo, useState } from 'react'
import { HandCoins, Plus, Search as SearchIcon } from 'lucide-react'
import { toast } from 'sonner'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { downloadBlob } from '@/features/billing/api/use-billing-ops'
import { useDebtsQuery, useDebtSummaryQuery, exportDebtsXlsx } from './api/use-debts'
import { DebtCreateDialog } from './components/debt-create-dialog'
import { DebtPayDialog } from './components/debt-pay-dialog'
import type { Debt } from '@/gen/v1/debt_pb'

function formatIDR(value: number | undefined) {
  return `Rp${Number(value || 0).toLocaleString('id-ID')}`
}

function formatDate(unix: number | bigint | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

const STATUS_LABEL: Record<string, string> = {
  OPEN: 'Belum Dibayar',
  PARTIAL: 'Dibayar Sebagian',
  PAID: 'Lunas',
  CANCELLED: 'Dibatalkan',
}

const STATUS_VARIANT: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  OPEN: 'destructive',
  PARTIAL: 'default',
  PAID: 'secondary',
  CANCELLED: 'outline',
}

export function DebtsPage() {
  const [status, setStatus] = useState('')
  const [search, setSearch] = useState('')
  const [overdueOnly, setOverdueOnly] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [payTarget, setPayTarget] = useState<Debt | null>(null)

  const filter = useMemo(
    () => ({ status, search, overdueOnly, limit: 200 }),
    [status, search, overdueOnly]
  )
  const { data, isLoading } = useDebtsQuery(filter)
  const { data: summary } = useDebtSummaryQuery()
  const { data: customers = [] } = useCustomersQuery()

  const customerMap = useMemo(() => {
    const m = new Map<string, string>()
    for (const c of customers) m.set(c.id, c.name)
    return m
  }, [customers])

  const debts = data?.debts ?? []

  async function handleExport() {
    try {
      const res = await exportDebtsXlsx(filter)
      downloadBlob(res.payload, res.contentType, res.filename)
    } catch (err) {
      toast.error(`Gagal ekspor: ${(err as Error).message}`)
    }
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <HandCoins className='h-6 w-6 text-primary' />
              <h1 className='text-2xl font-bold tracking-tight'>Hutang Pelanggan</h1>
            </div>
            <p className='mt-0.5 text-sm text-muted-foreground'>
              Catat hutang manual, terima pembayaran, dan pantau sisa tagihan
              pelanggan. Tanpa denda dan tanpa biaya tambahan.
            </p>
          </div>
          <div className='flex items-center gap-2'>
            <Button variant='outline' size='sm' onClick={handleExport}>
              Ekspor XLSX
            </Button>
            <Button size='sm' className='gap-1.5' onClick={() => setCreateOpen(true)}>
              <Plus className='h-4 w-4' /> Catat Hutang
            </Button>
          </div>
        </div>

        <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Total Hutang</CardDescription>
              <CardTitle className='text-xl'>{formatIDR(summary?.totalAmount)}</CardTitle>
            </CardHeader>
            <CardContent className='text-xs text-muted-foreground'>
              {summary?.count ?? 0} catatan hutang aktif
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Sisa Belum Dibayar</CardDescription>
              <CardTitle className='text-xl'>{formatIDR(summary?.outstanding)}</CardTitle>
            </CardHeader>
            <CardContent className='text-xs text-muted-foreground'>
              Belum termasuk hutang yang dibatalkan
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Melewati Jatuh Tempo</CardDescription>
              <CardTitle className='text-xl text-destructive'>
                {summary?.overdue ?? 0} hutang
              </CardTitle>
            </CardHeader>
            <CardContent className='text-xs text-muted-foreground'>
              Ditagih via pengingat, bukan denda
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Hasil Filter</CardDescription>
              <CardTitle className='text-xl'>{data?.total ?? 0} catatan</CardTitle>
            </CardHeader>
            <CardContent className='text-xs text-muted-foreground'>Maksimal 200 baris</CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader className='flex flex-row flex-wrap items-center justify-between gap-3 space-y-0'>
            <div>
              <CardTitle className='text-base'>Daftar Hutang</CardTitle>
              <CardDescription>Kelola pembayaran dan pantau sisa hutang.</CardDescription>
            </div>
            <div className='flex flex-wrap items-center gap-2'>
              <div className='relative'>
                <SearchIcon className='absolute left-2 top-2.5 h-4 w-4 text-muted-foreground' />
                <Input
                  className='h-9 w-56 pl-8'
                  placeholder='Cari pelanggan / deskripsi'
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <Select value={status || 'ALL'} onValueChange={(v) => setStatus(v === 'ALL' ? '' : v)}>
                <SelectTrigger className='h-9 w-44'>
                  <SelectValue placeholder='Semua status' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='ALL'>Semua status</SelectItem>
                  <SelectItem value='OPEN'>Belum Dibayar</SelectItem>
                  <SelectItem value='PARTIAL'>Dibayar Sebagian</SelectItem>
                  <SelectItem value='PAID'>Lunas</SelectItem>
                  <SelectItem value='CANCELLED'>Dibatalkan</SelectItem>
                </SelectContent>
              </Select>
              <label className='flex items-center gap-2 text-sm text-muted-foreground'>
                <Switch checked={overdueOnly} onCheckedChange={setOverdueOnly} />
                Jatuh tempo
              </label>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pelanggan</TableHead>
                  <TableHead>Deskripsi</TableHead>
                  <TableHead className='text-right'>Jumlah</TableHead>
                  <TableHead className='text-right'>Sisa</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Jatuh Tempo</TableHead>
                  <TableHead className='text-right'>Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell colSpan={7} className='py-8 text-center text-sm text-muted-foreground'>
                      Memuat hutang…
                    </TableCell>
                  </TableRow>
                ) : debts.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className='py-8 text-center text-sm text-muted-foreground'>
                      Belum ada hutang tercatat.
                    </TableCell>
                  </TableRow>
                ) : (
                  debts.map((d) => {
                    const outstanding = d.status === 'OPEN' || d.status === 'PARTIAL'
                    return (
                      <TableRow key={d.id}>
                        <TableCell className='font-medium'>
                          {customerMap.get(d.customerId) || d.customerId}
                        </TableCell>
                        <TableCell className='max-w-[220px] truncate text-muted-foreground'>
                          {d.description || '-'}
                        </TableCell>
                        <TableCell className='text-right'>{formatIDR(d.amount)}</TableCell>
                        <TableCell className='text-right font-medium'>
                          {formatIDR(d.remainingAmount)}
                        </TableCell>
                        <TableCell>
                          <Badge variant={STATUS_VARIANT[d.status] || 'outline'}>
                            {STATUS_LABEL[d.status] || d.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{formatDate(d.dueDateUnix)}</TableCell>
                        <TableCell className='text-right'>
                          {outstanding && (
                            <Button
                              variant='outline'
                              size='sm'
                              onClick={() => setPayTarget(d)}
                            >
                              Bayar
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </Main>

      {createOpen && (
        <DebtCreateDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          customers={customers}
        />
      )}
      {payTarget && (
        <DebtPayDialog
          open
          onOpenChange={(open) => !open && setPayTarget(null)}
          debt={payTarget}
          customerName={customerMap.get(payTarget.customerId) || ''}
        />
      )}
    </>
  )
}
