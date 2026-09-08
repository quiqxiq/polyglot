import { useMemo } from 'react'
import { AlertCircle, Repeat } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useSubscriptionsQuery } from '@/features/billing/api/use-billing'
import { SubscriptionsPrimaryButtons } from './components/subscriptions-primary-buttons'
import { SubscriptionsDialogs } from './components/subscriptions-dialogs'
import { SubscriptionsProvider } from './components/subscriptions-provider'
import { SubscriptionsTable } from './components/subscriptions-table'

export function Subscriptions() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { data: subscriptions = [], isLoading } = useSubscriptionsQuery('')

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const filteredSubscriptions = useMemo(() => {
    if (!selectedDeviceId) return []
    return subscriptions.filter((s) => s.deviceId === selectedDeviceId)
  }, [subscriptions, selectedDeviceId])

  return (
    <SubscriptionsProvider>
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
              <Repeat className='size-6 text-primary' />
              <h2 className='text-2xl font-bold tracking-tight'>Subscriptions</h2>
            </div>
            <p className='text-sm text-muted-foreground mt-0.5'>
              {currentDevice ? (
                <>
                  Router:{' '}
                  <span className='font-semibold text-foreground'>
                    {currentDevice.name}
                  </span>{' '}
                  ({currentDevice.host})
                </>
              ) : (
                'Pilih router di sidebar untuk mengelola langganan pelanggan.'
              )}
            </p>
          </div>
          <SubscriptionsPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Silakan pilih router MikroTik dari dropdown di sidebar untuk mengelola langganan pelanggan.
            </p>
          </div>
        ) : (
          <SubscriptionsTable data={filteredSubscriptions} isLoading={isLoading} />
        )}
      </Main>

      <SubscriptionsDialogs />
    </SubscriptionsProvider>
  )
}
