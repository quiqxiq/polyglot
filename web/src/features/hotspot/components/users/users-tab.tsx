import { useHotspotUsersQuery } from '../../api/use-hotspot-users'
import { useHotspotProfilesQuery } from '../../api/use-hotspot-profiles'
import { useDeviceStore } from '@/stores/device-store'
import { UsersTable } from './users-table'

export function UsersTab() {
  const { selectedDeviceId } = useDeviceStore()

  const { data: profiles = [] } = useHotspotProfilesQuery(selectedDeviceId)
  const { data: users = [], isLoading } = useHotspotUsersQuery(selectedDeviceId)

  return (
    <UsersTable
      data={users}
      profiles={profiles}
      isLoading={isLoading}
    />
  )
}

