import { useMemo } from 'react'
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
import { RolesTable } from './components/roles-table'
import { PoliciesTable } from './components/policies-table'

export function RBAC() {
  const { data: policies = [], isLoading: policiesLoading } = usePoliciesQuery()
  const { data: assignments = [] } = useRoleAssignmentsQuery()
  const { data: usersData, isLoading: usersLoading } = useUsersQuery()

  const availableRoles = useMemo(() => {
    const set = new Set<string>()
    policies.forEach((p) => {
      if (p.sub) set.add(p.sub)
    })
    assignments.forEach((a) => {
      if (a.role) set.add(a.role)
    })
    return Array.from(set)
  }, [policies, assignments])

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
            Manage roles, permission matrices, and user role assignments. Owner-only area.
          </p>
        </div>

        <Separator className='shadow-xs' />

        <Tabs defaultValue='roles' className='flex flex-1 flex-col'>
          <TabsList className='w-fit'>
            <TabsTrigger value='roles'>Roles &amp; Permissions</TabsTrigger>
            <TabsTrigger value='assignments'>Role Assignments</TabsTrigger>
            <TabsTrigger value='policies'>Raw Policies</TabsTrigger>
          </TabsList>
          <TabsContent value='roles' className='flex flex-1 flex-col'>
            <RolesTable policies={policies} isLoading={policiesLoading} />
          </TabsContent>
          <TabsContent value='assignments' className='flex flex-1 flex-col'>
            <AssignmentsTable
              users={usersData?.users ?? []}
              assignments={assignments}
              availableRoles={availableRoles}
              isLoading={usersLoading}
            />
          </TabsContent>
          <TabsContent value='policies' className='flex flex-1 flex-col'>
            <PoliciesTable policies={policies} isLoading={policiesLoading} />
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}
