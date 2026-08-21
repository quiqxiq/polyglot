import { Route } from '@/routes/_authenticated/hotspot/index'
import { Wifi, Users, PieChart, Activity, Radio, Laptop, ShieldCheck, Cookie, AlertCircle } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { HotspotProvider } from './context/hotspot-context'
import { HotspotPrimaryButtons } from './components/hotspot-primary-buttons'
import { HotspotDialogs } from './components/hotspot-dialogs'
import { UsersTab } from './components/users/users-tab'
import { ProfilesTab } from './components/profiles/profiles-tab'
import { ActiveTab } from './components/active/active-tab'
import { InactiveTab } from './components/inactive/inactive-tab'
import { HostsTab } from './components/hosts/hosts-tab'
import { BindingsTab } from './components/bindings/bindings-tab'
import { CookiesTab } from './components/cookies/cookies-tab'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'

type HotspotTabType = 'users' | 'profiles' | 'active' | 'inactive' | 'hosts' | 'bindings' | 'cookies'

function HotspotContent() {
  const navigate = Route.useNavigate()
  const search = Route.useSearch()
  const currentTab = (search.tab as HotspotTabType) || 'users'
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const handleTabChange = (val: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        tab: val as HotspotTabType,
      }),
      replace: true,
    })
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
        {/* ===== Top Title & Primary Actions ===== */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <Wifi className='size-6 text-primary' />
              <h1 className='text-2xl font-bold tracking-tight'>Hotspot Management</h1>
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
                'Select a router in the sidebar to view and manage hotspot users.'
              )}
            </p>
          </div>
          <HotspotPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Please select a MikroTik router from the top dropdown in the sidebar to start managing hotspot sessions.
            </p>
          </div>
        ) : (
          /* ===== Main Tabs ===== */
          <Tabs
            value={currentTab}
            onValueChange={handleTabChange}
            className='flex flex-1 flex-col gap-4'
          >
            <div className='w-full overflow-x-auto pb-1'>
              <TabsList className='h-10'>
                <TabsTrigger value='users' className='gap-2'>
                  <Users className='size-4' />
                  Users
                </TabsTrigger>
                <TabsTrigger value='profiles' className='gap-2'>
                  <PieChart className='size-4' />
                  User Profile
                </TabsTrigger>
                <TabsTrigger value='active' className='gap-2'>
                  <Activity className='size-4 text-emerald-500' />
                  Active
                </TabsTrigger>
                <TabsTrigger value='inactive' className='gap-2'>
                  <Radio className='size-4 text-sky-500' />
                  Inactive
                </TabsTrigger>
                <TabsTrigger value='hosts' className='gap-2'>
                  <Laptop className='size-4' />
                  Hosts
                </TabsTrigger>
                <TabsTrigger value='bindings' className='gap-2'>
                  <ShieldCheck className='size-4 text-emerald-500' />
                  IP Bindings
                </TabsTrigger>
                <TabsTrigger value='cookies' className='gap-2'>
                  <Cookie className='size-4 text-amber-500' />
                  Cookies
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value='users' className='flex flex-1 flex-col gap-4 m-0'>
              <UsersTab />
            </TabsContent>

            <TabsContent value='profiles' className='flex flex-1 flex-col gap-4 m-0'>
              <ProfilesTab />
            </TabsContent>

            <TabsContent value='active' className='flex flex-1 flex-col gap-4 m-0'>
              <ActiveTab />
            </TabsContent>

            <TabsContent value='inactive' className='flex flex-1 flex-col gap-4 m-0'>
              <InactiveTab />
            </TabsContent>

            <TabsContent value='hosts' className='flex flex-1 flex-col gap-4 m-0'>
              <HostsTab />
            </TabsContent>

            <TabsContent value='bindings' className='flex flex-1 flex-col gap-4 m-0'>
              <BindingsTab />
            </TabsContent>

            <TabsContent value='cookies' className='flex flex-1 flex-col gap-4 m-0'>
              <CookiesTab />
            </TabsContent>
          </Tabs>
        )}
      </Main>

      <HotspotDialogs />
    </>
  )
}

export function Hotspot() {
  return (
    <HotspotProvider>
      <HotspotContent />
    </HotspotProvider>
  )
}
