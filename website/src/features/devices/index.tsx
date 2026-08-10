import { useState } from 'react'
import {
  MixerHorizontalIcon,
  Cross2Icon,
} from '@radix-ui/react-icons'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { DevicesProvider, useDevicesContext } from './components/devices-provider'
import { DevicesPrimaryButtons } from './components/devices-primary-buttons'
import { DevicesCardGrid } from './components/devices-card-grid'
import { DevicesTable } from './components/devices-table'
import { DevicesDialogs } from './components/devices-dialogs'
import { useDevicesQuery } from './api/use-devices'

type VendorFilter = 'all' | 'mikrotik' | 'cisco' | 'huawei' | 'genieacs'

function DevicesContent() {
  const { data: devices = [], isLoading } = useDevicesQuery()
  const { viewMode } = useDevicesContext()

  const [searchTerm, setSearchTerm] = useState('')
  const [vendorFilter, setVendorFilter] = useState<VendorFilter>('all')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')

  const filteredDevices = devices
    .filter((device) => {
      if (vendorFilter !== 'all' && device.vendor.toLowerCase() !== vendorFilter) {
        return false
      }
      if (!searchTerm) return true
      const q = searchTerm.toLowerCase()
      return (
        device.name.toLowerCase().includes(q) ||
        device.host.toLowerCase().includes(q) ||
        device.vendor.toLowerCase().includes(q) ||
        device.driverType.toLowerCase().includes(q)
      )
    })
    .sort((a, b) => {
      if (sortOrder === 'asc') {
        return a.name.localeCompare(b.name)
      }
      return b.name.localeCompare(a.name)
    })

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* ===== Title & Top Toolbar ===== */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Device Inventory</h1>
            <p className='text-muted-foreground text-sm'>
              Manage your network routers, access points, and devices with real-time monitoring.
            </p>
          </div>
          <DevicesPrimaryButtons />
        </div>

        {/* ===== Filter & Search Controls (Matching Apps Feature) ===== */}
        <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
          <div className='flex flex-1 flex-wrap items-center gap-2'>
            <div className='relative w-full sm:w-64'>
              <Input
                placeholder='Filter devices by name, IP, or vendor...'
                className='h-9 pr-8 text-xs sm:text-sm'
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              {searchTerm && (
                <button
                  type='button'
                  onClick={() => setSearchTerm('')}
                  className='absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                  title='Clear search'
                >
                  <Cross2Icon className='h-3.5 w-3.5' />
                </button>
              )}
            </div>

            <Select value={vendorFilter} onValueChange={(val) => setVendorFilter(val as VendorFilter)}>
              <SelectTrigger className='h-9 w-36 text-xs sm:text-sm'>
                <SelectValue placeholder='All Vendors' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='all'>All Vendors</SelectItem>
                <SelectItem value='mikrotik'>MikroTik</SelectItem>
                <SelectItem value='cisco'>Cisco</SelectItem>
                <SelectItem value='huawei'>Huawei</SelectItem>
                <SelectItem value='genieacs'>GenieACS</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className='flex items-center gap-2 self-end sm:self-auto'>
            <Select value={sortOrder} onValueChange={(val) => setSortOrder(val as 'asc' | 'desc')}>
              <SelectTrigger className='h-9 w-28 text-xs sm:text-sm'>
                <div className='flex items-center gap-1.5'>
                  <MixerHorizontalIcon className='h-3.5 w-3.5' />
                  <span>{sortOrder === 'asc' ? 'Ascending' : 'Descending'}</span>
                </div>
              </SelectTrigger>
              <SelectContent align='end'>
                <SelectItem value='asc'>Ascending (A-Z)</SelectItem>
                <SelectItem value='desc'>Descending (Z-A)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <Separator className='shadow-xs' />

        {/* ===== Main View (Card Grid or Table) ===== */}
        {viewMode === 'card' ? (
          <DevicesCardGrid
            devices={filteredDevices}
            isLoading={isLoading}
            searchTerm={searchTerm}
          />
        ) : (
          <DevicesTable data={filteredDevices} isLoading={isLoading} />
        )}
      </Main>

      <DevicesDialogs />
    </>
  )
}

export function Devices() {
  return (
    <DevicesProvider>
      <DevicesContent />
    </DevicesProvider>
  )
}
