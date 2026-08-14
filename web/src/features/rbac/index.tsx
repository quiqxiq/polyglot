import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useUsersQuery } from '@/features/users/api/use-users'
import { usePoliciesQuery, useRoleAssignmentsQuery } from './api/use-rbac'
import { AssignmentsTable } from './components/assignments-table'
import { PoliciesTable } from './components/policies-table'

export function RBAC() {
  const { data: policies = [], isLoading: policiesLoading } = usePoliciesQuery()
  const { data: assignments = [] } = useRoleAssignmentsQuery()
  const { data: usersData, isLoading: usersLoading } = useUsersQuery()

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>
            Role-Based Access Control
          </h1>
          <p className='text-sm text-muted-foreground'>
            Manage role policies and user role assignments. Owner-only area.
          </p>
        </div>

        <Separator className='shadow-xs' />

        <Tabs defaultValue='policies' className='flex flex-1 flex-col'>
          <TabsList className='w-fit'>
            <TabsTrigger value='policies'>Policies</TabsTrigger>
            <TabsTrigger value='assignments'>Role Assignments</TabsTrigger>
          </TabsList>
          <TabsContent value='policies' className='flex flex-1 flex-col'>
            <PoliciesTable policies={policies} isLoading={policiesLoading} />
          </TabsContent>
          <TabsContent value='assignments' className='flex flex-1 flex-col'>
            <AssignmentsTable
              users={usersData?.users ?? []}
              assignments={assignments}
              isLoading={usersLoading}
            />
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}
