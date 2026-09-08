import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  PlanImportRow,
  PullRouterPlansRequest,
  CommitPlansRequest,
} from '@/gen/v1/ispadmin_pb'
import {
  ArrowLeft,
  CheckCircle2,
  Loader2,
  ArrowRight,
  ShieldCheck,
  RotateCcw,
} from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { formatAutoName } from '@/lib/format-name'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import {
  usePullRouterPlansMutation,
  useCommitPlansMutation,
} from '../api/use-plans'
import { PlansImportSourceTabs } from './plans-import-source-tabs'
import { PlansImportStats } from './plans-import-stats'
import { PlansImportTable } from './plans-import-table'

export function PlansImportPage() {
  const storeDeviceId = useDeviceStore((s) => s.selectedDeviceId)
  const [activeTab, setActiveTab] = useState<'router' | 'file'>('router')
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>(storeDeviceId || '')
  const [serviceType, setServiceType] = useState<string>('ALL')
  const [rows, setRows] = useState<PlanImportRow[]>([])
  const [commitSuccess, setCommitSuccess] = useState<{
    count: number
  } | null>(null)

  const { data: devices = [], isLoading: isDevicesLoading } = useDevicesQuery()
  const pullRouterPlans = usePullRouterPlansMutation()
  const commitPlans = useCommitPlansMutation()

  // ─── Handler Tarik Profil dari Router ────────────────────────────────────
  const handlePullRouter = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }

    try {
      const res = await pullRouterPlans.mutateAsync(
        new PullRouterPlansRequest({
          deviceId: selectedDeviceId,
          serviceType,
        })
      )

      const mappedRows = (res.rows || []).map((r) => {
        const item = new PlanImportRow(r)
        item.selected = true
        return item
      })

      setRows(mappedRows)
      setCommitSuccess(null)

      if (mappedRows.length === 0) {
        toast.info('Tidak ditemukan profile PPP atau Hotspot di router ini.')
      } else {
        toast.success(
          `Ditemukan ${mappedRows.length} profile (${res.pppoeDetected} PPPoE, ${res.hotspotDetected} Hotspot)`
        )
      }
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : 'Gagal menarik data profil dari router'
      )
    }
  }

  // ─── Table Row Callbacks ──────────────────────────────────────────────────
  const handleUpdateRow = (index: number, updated: Partial<PlanImportRow>) => {
    setRows((prev) => {
      const next = [...prev]
      const curr = next[index]
      if (curr) {
        next[index] = new PlanImportRow({ ...curr, ...updated })
      }
      return next
    })
  }

  const handleSelectOnlyNew = () => {
    setRows((prev) =>
      prev.map((r) => {
        const item = new PlanImportRow(r)
        item.selected = r.isNew
        return item
      })
    )
  }

  const handleBatchSetPrice = (price: number) => {
    setRows((prev) =>
      prev.map((r) => {
        if (r.selected) {
          const item = new PlanImportRow(r)
          item.price = price
          return item
        }
        return r
      })
    )
    toast.success(
      `Harga Rp ${price.toLocaleString('id-ID')} diterapkan ke paket terpilih`
    )
  }

  const handleBatchAutoFormatPlanNames = () => {
    const hasSelected = rows.some((r) => r.selected)
    let affected = 0
    setRows((prev) =>
      prev.map((r) => {
        if (hasSelected && !r.selected) return r
        const item = new PlanImportRow(r)
        item.name = formatAutoName(r.routerProfile || r.name)
        affected++
        return item
      })
    )
    toast.success(`Nama otomatis diformat untuk ${affected} paket`)
  }

  // ─── Commit Plans ────────────────────────────────────────────────────────
  const selectedRows = rows.filter((r) => r.selected)

  const handleCommit = async () => {
    if (selectedRows.length === 0) {
      toast.error('Pilih setidaknya 1 paket untuk disimpan')
      return
    }

    try {
      const res = await commitPlans.mutateAsync(
        new CommitPlansRequest({
          rows: selectedRows,
        })
      )

      setCommitSuccess({ count: res.plansCreated + res.plansUpdated })
      toast.success(
        `Berhasil menyimpan: ${res.plansCreated} paket dibuat, ${res.plansUpdated} paket diperbarui`
      )
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal menyimpan data paket'
      )
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

      <Main className='flex flex-1 flex-col gap-6 p-4 sm:p-6'>
        {/* Breadcrumb & Navigation */}
        <div className='flex items-center gap-2 text-sm text-muted-foreground'>
          <Link
            to='/plans'
            className='flex items-center gap-1 transition-colors hover:text-foreground'
          >
            <ArrowLeft className='h-4 w-4' />
            Kembali ke Service Plans
          </Link>
          <span>/</span>
          <span className='font-medium text-foreground'>Import</span>
        </div>

        {/* Page Title */}
        <div className='flex flex-wrap items-center justify-between gap-4'>
          <div>
            <div className='flex items-center gap-2'>
              <h1 className='text-2xl font-bold tracking-tight sm:text-3xl'>
                Import Service Plans
              </h1>
              <Badge
                variant='outline'
                className='border-primary/40 text-primary'
              >
                Dual-Method
              </Badge>
            </div>
            <p className='mt-1 max-w-3xl text-sm text-muted-foreground'>
              Tarik konfigurasi profile dari MikroTik (/ppp/profile &
              /ip/hotspot/user/profile), tinjau dan sesuaikan harga jual serta
              parameter jaringan, lalu simpan ke database.
            </p>
          </div>
        </div>

        {/* Mode Tabs */}
        <PlansImportSourceTabs
          activeTab={activeTab}
          onTabChange={setActiveTab}
          devices={devices}
          isDevicesLoading={isDevicesLoading}
          selectedDeviceId={selectedDeviceId}
          onSelectDevice={setSelectedDeviceId}
          serviceType={serviceType}
          onSelectServiceType={setServiceType}
          onPullRouter={handlePullRouter}
          isPulling={pullRouterPlans.isPending}
        />

        {/* KPI Summary Cards */}
        {rows.length > 0 && <PlansImportStats rows={rows} />}

        {/* Success Banner */}
        {commitSuccess && (
          <div className='flex flex-wrap items-center justify-between gap-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4'>
            <div className='flex items-center gap-3'>
              <CheckCircle2 className='h-6 w-6 text-emerald-600' />
              <div>
                <p className='font-medium text-emerald-800 dark:text-emerald-300'>
                  {commitSuccess.count} Service Plans Berhasil Disimpan!
                </p>
                <p className='text-xs text-muted-foreground'>
                  Paket layanan sudah aktif dan siap digunakan untuk pemetaan
                  pelanggan.
                </p>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                onClick={() => setCommitSuccess(null)}
              >
                <RotateCcw className='mr-1.5 h-3.5 w-3.5' />
                Import Lagi
              </Button>
              <Button size='sm' asChild>
                <Link to='/customers/import'>
                  <span>Lanjut ke Import Pelanggan</span>
                  <ArrowRight className='ml-1.5 h-3.5 w-3.5' />
                </Link>
              </Button>
            </div>
          </div>
        )}

        {/* Interactive TanStack Table View */}
        {rows.length > 0 && (
          <div className='space-y-4'>
            <PlansImportTable
              rows={rows}
              onUpdateRow={handleUpdateRow}
              onSelectOnlyNew={handleSelectOnlyNew}
              onBatchSetPrice={handleBatchSetPrice}
              onAutoFormatNames={handleBatchAutoFormatPlanNames}
            />

            {/* Bottom Commit Action Bar */}
            <div className='flex flex-wrap items-center justify-between gap-4 rounded-lg border bg-card p-4 shadow-xs'>
              <div className='text-xs text-muted-foreground'>
                <span className='font-semibold text-foreground'>
                  {selectedRows.length}
                </span>{' '}
                dari {rows.length} paket dipilih untuk disimpan ke database.
              </div>

              <div className='flex items-center gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => setRows([])}
                  disabled={commitPlans.isPending}
                >
                  Reset
                </Button>
                <Button
                  size='sm'
                  onClick={handleCommit}
                  disabled={selectedRows.length === 0 || commitPlans.isPending}
                  className='min-w-[140px]'
                >
                  {commitPlans.isPending ? (
                    <>
                      <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      <span>Menyimpan...</span>
                    </>
                  ) : (
                    <>
                      <ShieldCheck className='mr-2 h-4 w-4' />
                      <span>Simpan ({selectedRows.length}) Paket</span>
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>
        )}
      </Main>
    </>
  )
}
