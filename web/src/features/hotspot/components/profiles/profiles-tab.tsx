import { useHotspotProfilesQuery } from '../../api/use-hotspot-profiles'
import { useDeviceStore } from '@/stores/device-store'
import { ProfilesTable } from './profiles-table'

export function ProfilesTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: profiles = [], isLoading } = useHotspotProfilesQuery(selectedDeviceId)

  return <ProfilesTable data={profiles} isLoading={isLoading} />
}
