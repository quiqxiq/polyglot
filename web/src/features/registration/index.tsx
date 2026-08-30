import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useRegistrationsQuery } from './api/use-registration'
import { RegistrationDialogs } from './components/registration-dialogs'
import { RegistrationPrimaryButtons } from './components/registration-primary-buttons'
import { RegistrationProvider } from './components/registration-provider'
import { RegistrationTable } from './components/registration-table'

export function Registrations() {
  const { data: registrations = [], isLoading } = useRegistrationsQuery()

  return (
    <RegistrationProvider>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Pendaftaran Calon Pelanggan</h2>
            <p className='text-muted-foreground'>
              Pipeline pendaftaran, survei, jadwal pasang teknisi, dan aktivasi router BRAS
            </p>
          </div>
          <RegistrationPrimaryButtons />
        </div>
        <RegistrationTable data={registrations} isLoading={isLoading} />
      </Main>

      <RegistrationDialogs />
    </RegistrationProvider>
  )
}
