import { useMemo } from 'react'
import { AlertCircle, Contact } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useSubscriptionsQuery } from '@/features/billing/api/use-billing'
import { useCustomersQuery } from './api/use-customer'
import { CustomersDialogs } from './components/customers-dialogs'
import { CustomersPrimaryButtons } from './components/customers-primary-buttons'
import { CustomersProvider } from './components/customers-provider'
import { CustomersTable } from './components/customers-table'

export function Customers() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { data: customers = [], isLoading: isLoadingCustomers } = useCustomersQuery()
  const { data: subscriptions = [], isLoading: isLoadingSubs } = useSubscriptionsQuery('')

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const routerCustomerIds = useMemo(() => {
    if (!selectedDeviceId) return new Set<string>()
    return new Set(
      subscriptions
        .filter((s) => s.deviceId === selectedDeviceId)
        .map((s) => s.customerId)
    )
  }, [subscriptions, selectedDeviceId])

  const filteredCustomers = useMemo(() => {
    if (!selectedDeviceId) return []
    return customers.filter((c) => routerCustomerIds.has(c.id))
  }, [customers, routerCustomerIds, selectedDeviceId])

  const isLoading = isLoadingCustomers || isLoadingSubs

  return (
    <CustomersProvider>
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
              <Contact className='size-6 text-primary' />
              <h2 className='text-2xl font-bold tracking-tight'>Customers</h2>
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
                'Pilih router di sidebar untuk melihat pelanggan terdaftar.'
              )}
            </p>
          </div>
          <CustomersPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Silakan pilih router MikroTik dari dropdown di sidebar untuk melihat pelanggan.
            </p>
          </div>
        ) : (
          <CustomersTable data={filteredCustomers} isLoading={isLoading} />
        )}
      </Main>

      <CustomersDialogs />
    </CustomersProvider>
  )
}
