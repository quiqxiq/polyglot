import { Route } from '@/routes/_authenticated/ppp/index'
import {
  Activity,
  AlertCircle,
  KeyRound,
  Network,
  Shield,
  UserX,
} from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { PPPProvider } from './context/ppp-context'
import { PPPPrimaryButtons } from './components/ppp-primary-buttons'
import { PPPDialogs } from './components/ppp-dialogs'
import { SecretsTab } from './components/secrets/secrets-tab'
import { ActiveTab } from './components/active/active-tab'
import { InactiveTab } from './components/inactive/inactive-tab'
import { ProfilesTab } from './components/profiles/profiles-tab'

type PPPTabType = 'secrets' | 'active' | 'inactive' | 'profiles'

function PPPContent() {
  const navigate = Route.useNavigate()
  const search = Route.useSearch()
  const currentTab = (search.tab as PPPTabType) || 'secrets'
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const handleTabChange = (val: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        tab: val as PPPTabType,
      }),
      replace: true,
    })
  }

  return (
    <>
      <Header fixed>
        <Search className="me-auto" />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className="flex flex-1 flex-col gap-4 sm:gap-6">
        {/* ===== Top Title & Primary Actions ===== */}
        <div className="flex flex-wrap items-end justify-between gap-2">
          <div>
            <div className="flex items-center gap-2">
              <Network className="size-6 text-primary" />
              <h1 className="text-2xl font-bold tracking-tight">PPPoE & PPP Management</h1>
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">
              {currentDevice ? (
                <>
                  Router:{' '}
                  <span className="font-semibold text-foreground">
                    {currentDevice.name}
                  </span>{' '}
                  ({currentDevice.host})
                </>
              ) : (
                'Select a router to view and manage PPPoE subscribers, profiles, and active sessions.'
              )}
            </p>
          </div>
          <PPPPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className="flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed">
            <AlertCircle className="size-10 text-muted-foreground mb-3" />
            <h3 className="text-lg font-semibold">No Router Selected</h3>
            <p className="text-sm text-muted-foreground max-w-sm mt-1">
              Please select a MikroTik router from the top selector in the sidebar to start managing PPPoE subscribers and profiles.
            </p>
          </div>
        ) : (
          /* ===== Main Tabs ===== */
          <Tabs
            value={currentTab}
            onValueChange={handleTabChange}
            className="flex flex-1 flex-col gap-4"
          >
            <div className="w-full overflow-x-auto pb-1">
              <TabsList className="h-10">
                <TabsTrigger value="secrets" className="gap-2">
                  <KeyRound className="size-4" />
                  Secrets
                </TabsTrigger>
                <TabsTrigger value="active" className="gap-2">
                  <Activity className="size-4 text-emerald-500" />
                  Active
                </TabsTrigger>
                <TabsTrigger value="inactive" className="gap-2">
                  <UserX className="size-4 text-muted-foreground" />
                  Inactive
                </TabsTrigger>
                <TabsTrigger value="profiles" className="gap-2">
                  <Shield className="size-4 text-primary" />
                  Profiles
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="secrets" className="m-0 flex-1">
              <SecretsTab />
            </TabsContent>

            <TabsContent value="active" className="m-0 flex-1">
              <ActiveTab />
            </TabsContent>

            <TabsContent value="inactive" className="m-0 flex-1">
              <InactiveTab />
            </TabsContent>

            <TabsContent value="profiles" className="m-0 flex-1">
              <ProfilesTab />
            </TabsContent>
          </Tabs>
        )}

        <PPPDialogs />
      </Main>
    </>
  )
}

export function PPPFeature() {
  return (
    <PPPProvider>
      <PPPContent />
    </PPPProvider>
  )
}

export { PPPFeature as PPP }
export default PPPFeature

