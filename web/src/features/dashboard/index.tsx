import { useMemo } from 'react'
import { Link, useNavigate, useSearch } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { TopNav, type TopNavLink } from '@/components/layout/top-nav'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useWARealtimeStream } from '@/features/whatsapp/api/use-whatsapp-sse'
import { KPICards } from './components/kpi-cards'
import { SalesChart } from './components/sales-chart'
import { RecentVoucherSales } from './components/recent-voucher-sales'
import { DeviceFleetCard } from './components/device-fleet-card'
import { QuickActions } from './components/quick-actions'
import { PPPSubscriberSearchCard } from './components/ppp-subscriber-search-card'
import {
  ArrowUpRight,
  FileText,
  Network,
  RefreshCw,
  TrendingUp,
} from 'lucide-react'

type DashboardTab = 'utama' | 'laporan' | 'trend' | 'kpi' | 'lainnya'


export function Dashboard() {
  const queryClient = useQueryClient()
  const search = useSearch({ strict: false }) as { tab?: DashboardTab }
  const navigate = useNavigate()
  const currentTab: DashboardTab = search.tab || 'utama'
  const isUtama = currentTab === 'utama'

  const { data: devices = [] } = useDevicesQuery()
  const { selectedDeviceId } = useDeviceStore()
  useWARealtimeStream()

  // Gunakan router yang dipilih dari store, fallback jika belum ada
  const activeDeviceId = useMemo(() => {
    if (selectedDeviceId && devices.some((d) => d.id === selectedDeviceId)) {
      return selectedDeviceId
    }
    return devices[0]?.id || ''
  }, [selectedDeviceId, devices])

  const handleRefreshAll = () => {
    queryClient.invalidateQueries()
  }

  const handleTabChange = (tab: DashboardTab) => {
    navigate({
      to: '.',
      search: (prev: Record<string, unknown>) => ({
        ...prev,
        tab,
      }),
      replace: true,
    })
  }

  const topNavLinks: TopNavLink[] = [
    {
      title: 'Utama',
      href: '/',
      search: { tab: 'utama' },
      isActive: isUtama,
      onClick: () => handleTabChange('utama'),
      icon: <Network className='size-4' />,
    },
    {
      title: 'Laporan, Trend & KPI',
      href: '/',
      search: { tab: 'laporan' },
      isActive: !isUtama,
      onClick: () => handleTabChange('laporan'),
      icon: <TrendingUp className='size-4' />,
    },
  ]

  return (
    <>
      {/* ===== Header ===== */}
      <Header fixed>
        <TopNav links={topNavLinks} />
        <div className='ms-auto flex items-center space-x-2 sm:space-x-4'>
          <Search />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Main Content ===== */}
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Tab: Utama - Berisi HANYA PPPSubscriberSearchCard */}
        {isUtama ? (
          <>
            <div className='flex flex-wrap items-center justify-between gap-3'>
              <div>
                <h1 className='text-2xl font-bold tracking-tight'>Pencarian PPP</h1>
              </div>
              <div className='flex items-center gap-2'>
                <Button
                  size='sm'
                  variant='outline'
                  className='h-8 gap-1.5 text-xs'
                  onClick={handleRefreshAll}
                  title='Segarkan data pelanggan'
                >
                  <RefreshCw className='size-3.5' /> Segarkan
                </Button>
              </div>
            </div>

            <PPPSubscriberSearchCard deviceId={activeDeviceId} />
          </>
        ) : (
          /* Tab: Laporan, Trend, KPI, dan Lainnya Disatukan */
          <>
            <div className='flex flex-wrap items-center justify-between gap-3'>
              <div>
                <h1 className='text-2xl font-bold tracking-tight'>Laporan, Trend & KPI</h1>
                <p className='text-xs text-muted-foreground mt-0.5'>
                  Ringkasan indikator kinerja jaringan, penjualan voucher, tren omset, dan armada router.
                </p>
              </div>
              <div className='flex items-center gap-2'>
                <Button
                  size='sm'
                  variant='outline'
                  className='h-8 gap-1.5 text-xs'
                  onClick={handleRefreshAll}
                  title='Segarkan seluruh metrik dan laporan'
                >
                  <RefreshCw className='size-3.5' /> Segarkan
                </Button>
                <Button
                  size='sm'
                  className='h-8 gap-1.5 text-xs'
                  asChild
                >
                  <Link to='/reports'>
                    <FileText className='size-3.5' /> Buka Laporan Lengkap <ArrowUpRight className='size-3.5' />
                  </Link>
                </Button>
              </div>
            </div>

            {/* KPI Metrics */}
            <KPICards deviceId={activeDeviceId} />

            {/* Trend & Laporan & Lainnya (Visuals, Fleet & Quick Actions) */}
            <div className='grid grid-cols-1 gap-4 lg:grid-cols-7'>
              {/* Main Visuals: Trend Chart & Router Fleet (4 cols) */}
              <div className='col-span-1 space-y-4 lg:col-span-4'>
                <SalesChart deviceId={activeDeviceId} />
                <DeviceFleetCard />
              </div>

              {/* Quick Actions & Recent Voucher Sales (3 cols) */}
              <div className='col-span-1 space-y-4 lg:col-span-3'>
                <QuickActions />
                <RecentVoucherSales deviceId={activeDeviceId} />
              </div>
            </div>
          </>
        )}
      </Main>
    </>
  )
}

