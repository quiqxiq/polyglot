import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ConfigDrawer } from '@/components/config-drawer'
import { DevicesProvider } from './components/devices-provider'
import { DevicesPrimaryButtons } from './components/devices-primary-buttons'
import { DevicesTable } from './components/devices-table'
import { DevicesDialogs } from './components/devices-dialogs'
import { useDevicesQuery } from './api/use-devices'

export function Devices() {
  const { data: devices = [], isLoading } = useDevicesQuery()

  return (
    <DevicesProvider>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Device Inventory</h2>
            <p className='text-muted-foreground'>
              Manage your network routers, access points, and devices.
            </p>
          </div>
          <DevicesPrimaryButtons />
        </div>
        <DevicesTable data={devices} isLoading={isLoading} />
      </Main>

      <DevicesDialogs />
    </DevicesProvider>
  )
}
