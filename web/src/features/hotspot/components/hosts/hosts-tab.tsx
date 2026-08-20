import { useHotspotHostsQuery } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import { HostsTable } from './hosts-table'

export function HostsTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: hosts = [], isLoading } = useHotspotHostsQuery(selectedDeviceId)

  return <HostsTable data={hosts} isLoading={isLoading} />
}
