import { useEffect, useMemo } from 'react'
import { Route } from '@/routes/_authenticated/ppp/index'
import {
  Activity,
  AlertCircle,
  KeyRound,
  Network,
  Shield,
  ShieldAlert,
  UserX,
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
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useCan } from '@/hooks/use-can'
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

  const canSecrets = useCan('ppp:secret:read')
  const canActive = useCan('ppp:active:read')
  const canInactive = useCan('ppp:inactive:read')
  const canProfiles = useCan('ppp:profile:read')

  const accessibleTabs = useMemo(() => {
    const tabs: PPPTabType[] = []
    if (canSecrets) tabs.push('secrets')
    if (canActive) tabs.push('active')
    if (canInactive) tabs.push('inactive')
    if (canProfiles) tabs.push('profiles')
    return tabs
  }, [canSecrets, canActive, canInactive, canProfiles])

  const activeTab = accessibleTabs.includes(currentTab)
    ? currentTab
    : accessibleTabs[0] || 'secrets'

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
        tab: val as PPPTabType,
      }),
      replace: true,
    })
  }

  const { t } = useI18n()

  return (
    <>
      <Header fixed>
        <Search className="me-auto" />
        <LanguageSwitch />
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
              <h1 className="text-xl sm:text-2xl font-bold tracking-tight">{t('ppp.title')}</h1>
            </div>
            <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
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
        ) : accessibleTabs.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed">
            <ShieldAlert className="size-10 text-destructive mb-3" />
            <h3 className="text-lg font-semibold">Access Denied</h3>
            <p className="text-sm text-muted-foreground max-w-sm mt-1">
              You do not have permission to view any PPP sub-tabs on this router.
            </p>
          </div>
        ) : (
          /* ===== Main Tabs ===== */
          <Tabs
            value={activeTab}
            onValueChange={handleTabChange}
            className="flex flex-1 flex-col gap-4"
          >
            <div className="w-full overflow-x-auto pb-1">
              <TabsList className="h-10">
                {canSecrets && (
                  <TabsTrigger value="secrets" className="gap-2">
                    <KeyRound className="size-4" />
                    Secrets
                  </TabsTrigger>
                )}
                {canActive && (
                  <TabsTrigger value="active" className="gap-2">
                    <Activity className="size-4 text-emerald-500" />
                    Active
                  </TabsTrigger>
                )}
                {canInactive && (
                  <TabsTrigger value="inactive" className="gap-2">
                    <UserX className="size-4 text-muted-foreground" />
                    Inactive
                  </TabsTrigger>
                )}
                {canProfiles && (
                  <TabsTrigger value="profiles" className="gap-2">
                    <Shield className="size-4 text-primary" />
                    Profiles
                  </TabsTrigger>
                )}
              </TabsList>
            </div>

            {canSecrets && (
              <TabsContent value="secrets" className="m-0 flex-1">
                <SecretsTab />
              </TabsContent>
            )}

            {canActive && (
              <TabsContent value="active" className="m-0 flex-1">
                <ActiveTab />
              </TabsContent>
            )}

            {canInactive && (
              <TabsContent value="inactive" className="m-0 flex-1">
                <InactiveTab />
              </TabsContent>
            )}

            {canProfiles && (
              <TabsContent value="profiles" className="m-0 flex-1">
                <ProfilesTab />
              </TabsContent>
            )}
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

