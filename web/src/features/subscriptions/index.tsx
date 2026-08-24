import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useSubscriptionsQuery } from '@/features/billing/api/use-billing'
import { SubscriptionsDialogs } from './components/subscriptions-dialogs'
import { SubscriptionsProvider } from './components/subscriptions-provider'
import { SubscriptionsTable } from './components/subscriptions-table'

export function Subscriptions() {
  const { data: subscriptions = [], isLoading } = useSubscriptionsQuery('')

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
            <h2 className='text-2xl font-bold tracking-tight'>Subscriptions</h2>
            <p className='text-muted-foreground'>
              Langganan pelanggan — paket aktif, isolir, penangguhan, dan provisi
              router
            </p>
          </div>
        </div>
        <SubscriptionsTable data={subscriptions} isLoading={isLoading} />
      </Main>

      <SubscriptionsDialogs />
    </SubscriptionsProvider>
  )
}
