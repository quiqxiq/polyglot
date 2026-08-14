import { useEffect, useState } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useUsersQuery } from './api/use-users'
import { UsersDialogs } from './components/users-dialogs'
import { UsersPrimaryButtons } from './components/users-primary-buttons'
import { UsersProvider } from './components/users-provider'
import { UsersTable } from './components/users-table'

function UsersContent() {
  const [searchTerm, setSearchTerm] = useState('')
  const [debounced, setDebounced] = useState('')

  // Debounce pencarian server-side (ListUsers search username/email).
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(searchTerm.trim()), 300)
    return () => clearTimeout(timer)
  }, [searchTerm])

  const { data, isLoading } = useUsersQuery(debounced)
  const users = data?.users ?? []
  const total = Number(data?.total ?? 0)

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>User List</h1>
            <p className='text-sm text-muted-foreground'>
              Manage users and their roles. {total} account
              {total === 1 ? '' : 's'}.
            </p>
          </div>
          <UsersPrimaryButtons />
        </div>

        <div className='flex flex-col gap-3 sm:flex-row sm:items-center'>
          <div className='relative w-full sm:w-72'>
            <Input
              placeholder='Filter by username or email...'
              className='h-9 pr-8 text-xs sm:text-sm'
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
            {searchTerm && (
              <button
                type='button'
                onClick={() => setSearchTerm('')}
                className='absolute top-1/2 right-2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                title='Clear search'
              >
                <Cross2Icon className='h-3.5 w-3.5' />
              </button>
            )}
          </div>
        </div>

        <Separator className='shadow-xs' />

        <UsersTable data={users} isLoading={isLoading} />
      </Main>

      <UsersDialogs />
    </>
  )
}

export function Users() {
  return (
    <UsersProvider>
      <UsersContent />
    </UsersProvider>
  )
}
