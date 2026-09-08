import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  CustomerSubscriptionImportRow,
  PullRouterCustomersRequest,
  CommitCustomersRequest,
} from '@/gen/v1/ispadmin_pb'
import {
  ArrowLeft,
  CheckCircle2,
  Loader2,
  ShieldCheck,
  RotateCcw,
  ArrowRight,
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
import { formatAutoName, parseUsernameNameAndAddress } from '@/lib/format-name'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import {
  usePullRouterCustomersMutation,
  useCommitCustomersMutation,
} from '../api/use-customer'
import { CustomersImportBatchBar } from './customers-import-batch-bar'
import { CustomersImportSourceTabs } from './customers-import-source-tabs'
import { CustomersImportStats } from './customers-import-stats'
import { CustomersImportTable } from './customers-import-table'

export function CustomersImportPage() {
  const [activeTab, setActiveTab] = useState<'router' | 'file'>('router')

  // ─── Filter & Options State ───────────────────────────────────────────────
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>('')
  const [pullPPPoE, setPullPPPoE] = useState<boolean>(true)
  const [pullHotspotMember, setPullHotspotMember] = useState<boolean>(true)
  const [pullHotspotIPBinding, setPullHotspotIPBinding] =
    useState<boolean>(true)
  const [pullHotspotVoucher, setPullHotspotVoucher] = useState<boolean>(false)

  // ─── Data & Mutation State ───────────────────────────────────────────────
  const [rows, setRows] = useState<CustomerSubscriptionImportRow[]>([])
  const [stats, setStats] = useState<{
    pppoe: number
    hotspot: number
    ipBinding: number
    vouchersSkipped: number
  } | null>(null)
  const [commitResult, setCommitResult] = useState<{
    customersCreated: number
    customersUpdated: number
    subscriptionsCreated: number
  } | null>(null)

  const { data: devices = [], isLoading: isDevicesLoading } = useDevicesQuery()
  const { data: plans = [] } = usePlansQuery(false)
  const pullCustomers = usePullRouterCustomersMutation()
  const commitCustomers = useCommitCustomersMutation()

  const computeServiceType = () => {
    const hasPPPoE = pullPPPoE
    const hasHotspot =
      pullHotspotMember || pullHotspotIPBinding || pullHotspotVoucher
    if (hasPPPoE && hasHotspot) return 'ALL'
    if (hasPPPoE) return 'PPPOE'
    if (hasHotspot) return 'HOTSPOT'
    return 'ALL'
  }

  // ─── Handler Tarik Akun dari Router ──────────────────────────────────────
  const handlePullRouter = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }

    try {
      const res = await pullCustomers.mutateAsync(
        new PullRouterCustomersRequest({
          deviceId: selectedDeviceId,
          serviceType: computeServiceType(),
          includeIpBindings: pullHotspotIPBinding,
          includeVouchers: pullHotspotVoucher,
        })
      )

      // Auto-match plans by planName if planId is empty
      const mappedRows = (res.rows || []).map((r) => {
        const item = new CustomerSubscriptionImportRow(r)
        item.selected = true

        if (!item.planId && item.planName) {
          const match = plans.find(
            (p) => p.name.toLowerCase() === item.planName.toLowerCase()
          )
          if (match) {
            item.planId = match.id
            if (item.price === 0) item.price = match.price
          }
        }
        return item
      })

      setRows(mappedRows)
      setStats({
        pppoe: res.pppoeDetected,
        hotspot: res.hotspotPermanentDetected,
        ipBinding: res.hotspotIpBindingDetected,
        vouchersSkipped: res.vouchersSkipped,
      })
      setCommitResult(null)

      if (mappedRows.length === 0) {
        toast.info(
          'Tidak ditemukan akun aktif di router ini dengan filter yang dipilih.'
        )
      } else {
        toast.success(
          `Ditemukan ${mappedRows.length} akun aktif (${res.pppoeDetected} PPPoE, ${res.hotspotPermanentDetected} Hotspot, ${res.hotspotIpBindingDetected} IP Binding)`
        )
      }
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : 'Gagal menarik data pelanggan dari router'
      )
    }
  }

  // ─── Row & Batch Callbacks ────────────────────────────────────────────────
  const handleUpdateRow = (
    index: number,
    updated: Partial<CustomerSubscriptionImportRow>
  ) => {
    setRows((prev) => {
      const next = [...prev]
      const curr = next[index]
      if (curr) {
        next[index] = new CustomerSubscriptionImportRow({ ...curr, ...updated })
      }
      return next
    })
  }

  const handleToggleSelectAll = (select: boolean) => {
    setRows((prev) =>
      prev.map((r) => {
        const item = new CustomerSubscriptionImportRow(r)
        item.selected = select
        return item
      })
    )
  }

  const handleBatchApplyPlan = (
    planId: string,
    planName: string,
    price: number
  ) => {
    setRows((prev) =>
      prev.map((r) => {
        if (r.selected) {
          const item = new CustomerSubscriptionImportRow(r)
          item.planId = planId
          item.planName = planName
          item.price = price
          return item
        }
        return r
      })
    )
    toast.success(`Paket "${planName}" diterapkan ke semua akun terpilih`)
  }

  const handleBatchApplyBillingDay = (day: number) => {
    setRows((prev) =>
      prev.map((r) => {
        if (r.selected) {
          const item = new CustomerSubscriptionImportRow(r)
          item.billingDay = day
          return item
        }
        return r
      })
    )
    toast.success(`Jatuh tempo tgl ${day} diterapkan ke semua akun terpilih`)
  }

  const handleBatchApplyDefaultAddress = (address: string) => {
    setRows((prev) =>
      prev.map((r) => {
        if (r.selected) {
          const item = new CustomerSubscriptionImportRow(r)
          item.address = address
          return item
        }
        return r
      })
    )
    toast.success('Alamat massal berhasil diterapkan')
  }

  const handleBatchAutoGenerateNames = (
    mode: 'name_only' | 'split_address' = 'name_only'
  ) => {
    const hasSelected = rows.some((r) => r.selected)
    let affected = 0
    setRows((prev) =>
      prev.map((r) => {
        if (hasSelected && !r.selected) return r
        const item = new CustomerSubscriptionImportRow(r)
        if (mode === 'split_address') {
          const { name, address } = parseUsernameNameAndAddress(r.username)
          item.name = name
          if (
            address &&
            (!item.address || item.address.startsWith('Area Router'))
          ) {
            item.address = address
          }
        } else {
          item.name = formatAutoName(r.username)
        }
        affected++
        return item
      })
    )
    toast.success(
      mode === 'split_address'
        ? `Nama & alamat di-generate untuk ${affected} akun`
        : `Nama otomatis di-generate untuk ${affected} akun`
    )
  }

  // ─── Commit Import ────────────────────────────────────────────────────────
  const selectedRows = rows.filter((r) => r.selected)

  const handleCommit = async () => {
    if (selectedRows.length === 0) {
      toast.error('Pilih setidaknya 1 akun untuk disimpan')
      return
    }

    try {
      const res = await commitCustomers.mutateAsync(
        new CommitCustomersRequest({
          deviceId: selectedDeviceId,
          rows: selectedRows,
        })
      )

      setCommitResult({
        customersCreated: res.customersCreated,
        customersUpdated: res.customersUpdated,
        subscriptionsCreated: res.subscriptionsCreated,
      })

      toast.success(
        `Import Selesai: ${res.customersCreated} pelanggan baru dibuat, ${res.subscriptionsCreated} langganan aktif terhubung.`
      )
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : 'Gagal menyimpan data import pelanggan'
      )
    }
  }

  const selectedDevice = devices.find((d) => d.id === selectedDeviceId)
  const readyCount = rows.filter((r) => r.phone && r.address && r.name).length

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-6 p-4 sm:p-6'>
        {/* Breadcrumb */}
        <div className='flex items-center gap-2 text-sm text-muted-foreground'>
          <Link
            to='/customers'
            className='flex items-center gap-1 transition-colors hover:text-foreground'
          >
            <ArrowLeft className='h-4 w-4' />
            Kembali ke Pelanggan
          </Link>
          <span>/</span>
          <span className='font-medium text-foreground'>
            Import Pelanggan & Langganan
          </span>
        </div>

        {/* Page Title */}
        <div className='flex flex-wrap items-center justify-between gap-4'>
          <div>
            <div className='flex items-center gap-2'>
              <h1 className='text-2xl font-bold tracking-tight sm:text-3xl'>
                Import Pelanggan & Langganan
              </h1>
              <Badge
                variant='outline'
                className='border-primary/40 text-primary'
              >
                Dual-Method
              </Badge>
            </div>
            <p className='mt-1 max-w-3xl text-sm text-muted-foreground'>
              Tarik akun aktif dari router MikroTik (PPPoE Secrets, Hotspot
              Members, IP Bindings) atau file spreadsheet. Sesuaikan data
              demografis (Nama, HP/WA, Alamat), tentukan paket layanan & harga,
              lalu simpan ke database.
            </p>
          </div>
        </div>

        {/* Mode Tabs */}
        <CustomersImportSourceTabs
          activeTab={activeTab}
          onTabChange={setActiveTab}
          devices={devices}
          isDevicesLoading={isDevicesLoading}
          selectedDeviceId={selectedDeviceId}
          onSelectDevice={setSelectedDeviceId}
          pullPPPoE={pullPPPoE}
          onTogglePPPoE={setPullPPPoE}
          pullHotspotMember={pullHotspotMember}
          onToggleHotspotMember={setPullHotspotMember}
          pullHotspotIPBinding={pullHotspotIPBinding}
          onToggleHotspotIPBinding={setPullHotspotIPBinding}
          pullHotspotVoucher={pullHotspotVoucher}
          onToggleHotspotVoucher={setPullHotspotVoucher}
          onPullRouter={handlePullRouter}
          isPulling={pullCustomers.isPending}
        />

        {/* Stat Cards */}
        {stats && rows.length > 0 && (
          <CustomersImportStats
            total={rows.length}
            pppoe={stats.pppoe}
            hotspot={stats.hotspot}
            ipBinding={stats.ipBinding}
            vouchersSkipped={stats.vouchersSkipped}
            readyCount={readyCount}
          />
        )}

        {/* Commit Success Banner */}
        {commitResult && (
          <div className='flex flex-wrap items-center justify-between gap-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4'>
            <div className='flex items-center gap-3'>
              <CheckCircle2 className='h-6 w-6 text-emerald-600' />
              <div>
                <p className='font-medium text-emerald-800 dark:text-emerald-300'>
                  Data Pelanggan & Langganan Berhasil Disimpan!
                </p>
                <p className='text-xs text-muted-foreground'>
                  {commitResult.customersCreated} pelanggan baru dibuat,{' '}
                  {commitResult.customersUpdated} diperbarui, dan{' '}
                  {commitResult.subscriptionsCreated} langganan aktif terhubung
                  ke router{' '}
                  <span className='font-semibold text-foreground'>
                    {selectedDevice?.name || 'MikroTik'}
                  </span>
                  .
                </p>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                onClick={() => setCommitResult(null)}
              >
                <RotateCcw className='mr-1.5 h-3.5 w-3.5' />
                Import Lagi
              </Button>
              <Button size='sm' asChild>
                <Link to='/customers'>
                  <span>Lihat Daftar Pelanggan</span>
                  <ArrowRight className='ml-1.5 h-3.5 w-3.5' />
                </Link>
              </Button>
            </div>
          </div>
        )}

        {/* Preview Table & Actions */}
        {rows.length > 0 && (
          <div className='space-y-4'>
            {/* Batch Toolbar */}
            <CustomersImportBatchBar
              plans={plans}
              selectedCount={selectedRows.length}
              totalCount={rows.length}
              allSelected={
                rows.length > 0 && selectedRows.length === rows.length
              }
              onApplyPlan={handleBatchApplyPlan}
              onApplyBillingDay={handleBatchApplyBillingDay}
              onApplyDefaultAddress={handleBatchApplyDefaultAddress}
              onAutoGenerateNames={handleBatchAutoGenerateNames}
              onToggleSelectAll={handleToggleSelectAll}
            />

            {/* Interactive TanStack Table */}
            <CustomersImportTable
              rows={rows}
              plans={plans}
              onUpdateRow={handleUpdateRow}
            />

            {/* Bottom Commit Action Bar */}
            <div className='flex flex-wrap items-center justify-between gap-4 rounded-lg border bg-card p-4 shadow-xs'>
              <div className='text-xs text-muted-foreground'>
                <span className='font-semibold text-foreground'>
                  {selectedRows.length}
                </span>{' '}
                dari {rows.length} akun dipilih untuk disimpan ke database.
              </div>

              <div className='flex items-center gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => setRows([])}
                  disabled={commitCustomers.isPending}
                >
                  Reset
                </Button>
                <Button
                  size='sm'
                  onClick={handleCommit}
                  disabled={
                    selectedRows.length === 0 || commitCustomers.isPending
                  }
                  className='min-w-[150px]'
                >
                  {commitCustomers.isPending ? (
                    <>
                      <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      <span>Menyimpan...</span>
                    </>
                  ) : (
                    <>
                      <ShieldCheck className='mr-2 h-4 w-4' />
                      <span>Simpan ({selectedRows.length}) Akun</span>
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
