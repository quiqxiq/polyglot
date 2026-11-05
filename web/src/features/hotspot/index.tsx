import { useEffect, useMemo } from 'react'
import { Route } from '@/routes/_authenticated/hotspot/index'
import {
  Wifi,
  Users,
  PieChart,
  Activity,
  Radio,
  Laptop,
  ShieldCheck,
  ShieldAlert,
  Cookie,
  AlertCircle,
} from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { LanguageSwitch } from '@/components/language-switch'
import { useI18n } from '@/context/i18n-provider'
import { useCan } from '@/hooks/use-can'
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

  const canUsers = useCan('hotspot:user:read')
  const canProfiles = useCan('hotspot:profile:read')
  const canActive = useCan('hotspot:active:read')
  const canHosts = useCan('hotspot:host:read')
  const canBindings = useCan('hotspot:binding:read')

  const accessibleTabs = useMemo(() => {
    const tabs: HotspotTabType[] = []
    if (canUsers) tabs.push('users')
    if (canProfiles) tabs.push('profiles')
    if (canActive) {
      tabs.push('active')
      tabs.push('inactive')
    }
    if (canHosts) tabs.push('hosts')
    if (canBindings) {
      tabs.push('bindings')
      tabs.push('cookies')
    }
    return tabs
  }, [canUsers, canProfiles, canActive, canHosts, canBindings])

  const activeTab = accessibleTabs.includes(currentTab)
    ? currentTab
    : accessibleTabs[0] || 'users'

  useEffect(() => {
    if (accessibleTabs.length > 0 && !accessibleTabs.includes(currentTab)) {
      navigate({
        search: (prev) => ({
          ...prev,
          tab: accessibleTabs[0],
        }),
        replace: true,
      })
    }
  }, [accessibleTabs, currentTab, navigate])

  const handleTabChange = (val: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        tab: val as HotspotTabType,
      }),
      replace: true,
    })
  }

  const { t } = useI18n()

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <LanguageSwitch />
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
              <h1 className='text-2xl font-bold tracking-tight'>{t('hotspot.title')}</h1>
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
        ) : accessibleTabs.length === 0 ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <ShieldAlert className='size-10 text-destructive mb-3' />
            <h3 className='text-lg font-semibold'>Access Denied</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              You do not have permission to view any Hotspot sub-tabs on this router.
            </p>
          </div>
        ) : (
          /* ===== Main Tabs ===== */
          <Tabs
            value={activeTab}
            onValueChange={handleTabChange}
            className='flex flex-1 flex-col gap-4'
          >
            <div className='w-full overflow-x-auto pb-1'>
              <TabsList className='h-10'>
                {canUsers && (
                  <TabsTrigger value='users' className='gap-2'>
                    <Users className='size-4' />
                    Users
                  </TabsTrigger>
                )}
                {canProfiles && (
                  <TabsTrigger value='profiles' className='gap-2'>
                    <PieChart className='size-4' />
                    User Profile
                  </TabsTrigger>
                )}
                {canActive && (
                  <>
                    <TabsTrigger value='active' className='gap-2'>
                      <Activity className='size-4 text-emerald-500' />
                      Active
                    </TabsTrigger>
                    <TabsTrigger value='inactive' className='gap-2'>
                      <Radio className='size-4 text-sky-500' />
                      Inactive
                    </TabsTrigger>
                  </>
                )}
                {canHosts && (
                  <TabsTrigger value='hosts' className='gap-2'>
                    <Laptop className='size-4' />
                    Hosts
                  </TabsTrigger>
                )}
                {canBindings && (
                  <>
                    <TabsTrigger value='bindings' className='gap-2'>
                      <ShieldCheck className='size-4 text-emerald-500' />
                      IP Bindings
                    </TabsTrigger>
                    <TabsTrigger value='cookies' className='gap-2'>
                      <Cookie className='size-4 text-amber-500' />
                      Cookies
                    </TabsTrigger>
                  </>
                )}
              </TabsList>
            </div>

            {canUsers && (
              <TabsContent value='users' className='flex flex-1 flex-col gap-4 m-0'>
                <UsersTab />
              </TabsContent>
            )}

            {canProfiles && (
              <TabsContent value='profiles' className='flex flex-1 flex-col gap-4 m-0'>
                <ProfilesTab />
              </TabsContent>
            )}

            {canActive && (
              <>
                <TabsContent value='active' className='flex flex-1 flex-col gap-4 m-0'>
                  <ActiveTab />
                </TabsContent>

                <TabsContent value='inactive' className='flex flex-1 flex-col gap-4 m-0'>
                  <InactiveTab />
                </TabsContent>
              </>
            )}

            {canHosts && (
              <TabsContent value='hosts' className='flex flex-1 flex-col gap-4 m-0'>
                <HostsTab />
              </TabsContent>
            )}

            {canBindings && (
              <>
                <TabsContent value='bindings' className='flex flex-1 flex-col gap-4 m-0'>
                  <BindingsTab />
                </TabsContent>

                <TabsContent value='cookies' className='flex flex-1 flex-col gap-4 m-0'>
                  <CookiesTab />
                </TabsContent>
              </>
            )}
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
