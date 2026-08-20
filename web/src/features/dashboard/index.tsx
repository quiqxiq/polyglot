import { useEffect, useMemo } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useWARealtimeStream } from '@/features/whatsapp/api/use-whatsapp-sse'
import { KPICards } from './components/kpi-cards'
import { SalesChart } from './components/sales-chart'
import { RecentVoucherSales } from './components/recent-voucher-sales'
import { DeviceFleetCard } from './components/device-fleet-card'
import { QuickActions } from './components/quick-actions'
import { RefreshCw } from 'lucide-react'

export function Dashboard() {
  const queryClient = useQueryClient()
  const { data: devices = [] } = useDevicesQuery()
  const { selectedDeviceId, setSelectedDeviceId } = useDeviceStore()
  useWARealtimeStream()

  // Fallback ke router pertama bila belum ada yang terpilih
  const activeDeviceId = useMemo(() => {
    if (selectedDeviceId && devices.some((d) => d.id === selectedDeviceId)) {
      return selectedDeviceId
    }
    return devices[0]?.id || ''
  }, [selectedDeviceId, devices])

  useEffect(() => {
    if (!selectedDeviceId && devices.length > 0) {
      setSelectedDeviceId(devices[0].id)
    }
  }, [selectedDeviceId, devices, setSelectedDeviceId])

  const handleRefreshAll = () => {
    queryClient.invalidateQueries()
  }

  return (
    <>
      {/* ===== Header ===== */}
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      {/* ===== Main Content ===== */}
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Title Bar */}
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Dashboard</h1>
            <p className='text-xs text-muted-foreground mt-0.5'>
              Ringkasan operasional jaringan, hotspot, dan layanan pelanggan.
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
          </div>
        </div>

        {/* Top 4 KPI Metrics */}
        <KPICards deviceId={activeDeviceId} />

        {/* Middle Section: Charts & Side widgets */}
        <div className='grid grid-cols-1 gap-4 lg:grid-cols-7'>
          {/* Main Visuals (Kolom Kiri 4 cols) */}
          <div className='col-span-1 space-y-4 lg:col-span-4'>
            <SalesChart deviceId={activeDeviceId} />
            <DeviceFleetCard />
          </div>

          {/* Quick Actions & Feed (Kolom Kanan 3 cols) */}
          <div className='col-span-1 space-y-4 lg:col-span-3'>
            <QuickActions />
            <RecentVoucherSales deviceId={activeDeviceId} />
          </div>
        </div>
      </Main>
    </>
  )
}
