import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { usePlansQuery } from './api/use-plans'
import { PlansDialogs } from './components/plans-dialogs'
import { PlansPrimaryButtons } from './components/plans-primary-buttons'
import { PlansProvider } from './components/plans-provider'
import { PlansTable } from './components/plans-table'

export function Plans() {
  const { data: plans = [], isLoading } = usePlansQuery(false)

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
            <h2 className='text-2xl font-bold tracking-tight'>Service Plans</h2>
            <p className='text-muted-foreground'>
              Paket layanan ISP — PPPoE, Hotspot, Dedicated dengan parameter
              MikroTik
            </p>
          </div>
          <PlansPrimaryButtons />
        </div>
        <PlansTable data={plans} isLoading={isLoading} />
      </Main>

      <PlansDialogs />
    </PlansProvider>
  )
}
