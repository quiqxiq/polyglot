import { useMemo } from 'react'
import { AlertCircle, Package } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { usePPPProfilesQuery } from '@/features/ppp/api/use-ppp-profiles'
import { useHotspotProfilesQuery } from '@/features/hotspot/api/use-hotspot-profiles'
import { useSubscriptionsQuery } from '@/features/billing/api/use-billing'
import { usePlansQuery } from './api/use-plans'
import { PlansDialogs } from './components/plans-dialogs'
import { PlansPrimaryButtons } from './components/plans-primary-buttons'
import { PlansProvider } from './components/plans-provider'
import { PlansTable } from './components/plans-table'

export function Plans() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { data: plans = [], isLoading: isLoadingPlans } = usePlansQuery(false)
  const { data: pppProfiles = [], isLoading: isLoadingPPP } = usePPPProfilesQuery(selectedDeviceId)
  const { data: hotspotProfiles = [], isLoading: isLoadingHotspot } = useHotspotProfilesQuery(selectedDeviceId)
  const { data: subscriptions = [], isLoading: isLoadingSubs } = useSubscriptionsQuery('')

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const routerProfileNames = useMemo(() => {
    const set = new Set<string>()
    for (const p of pppProfiles) {
      if (p.name) set.add(p.name.trim().toLowerCase())
    }
    for (const p of hotspotProfiles) {
      if (p.name) set.add(p.name.trim().toLowerCase())
    }
    return set
  }, [pppProfiles, hotspotProfiles])

  const routerSubPlanIds = useMemo(() => {
    if (!selectedDeviceId) return new Set<string>()
    return new Set(
      subscriptions
        .filter((s) => s.deviceId === selectedDeviceId && s.planId)
        .map((s) => s.planId)
    )
  }, [subscriptions, selectedDeviceId])

  const filteredPlans = useMemo(() => {
    if (!selectedDeviceId) return []
    return plans.filter((plan) => {
      if (routerProfileNames.has(plan.name.trim().toLowerCase())) return true
      if (routerSubPlanIds.has(plan.id)) return true
      return false
    })
  }, [plans, routerProfileNames, routerSubPlanIds, selectedDeviceId])

  const isLoading = isLoadingPlans || isLoadingPPP || isLoadingHotspot || isLoadingSubs

  return (
    <PlansProvider>
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
              <Package className='size-6 text-primary' />
              <h2 className='text-2xl font-bold tracking-tight'>Service Plans</h2>
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
                'Pilih router di sidebar untuk mengelola paket layanan.'
              )}
            </p>
          </div>
          <PlansPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Silakan pilih router MikroTik dari dropdown di sidebar untuk mengelola paket layanan.
            </p>
          </div>
        ) : (
          <PlansTable data={filteredPlans} isLoading={isLoading} />
        )}
      </Main>

      <PlansDialogs />
    </PlansProvider>
  )
}
