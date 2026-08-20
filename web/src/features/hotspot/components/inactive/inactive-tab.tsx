import { useStreamHotspotInactive } from '../../api/use-hotspot-stream'
import { useDeviceStore } from '@/stores/device-store'
import { InactiveTable } from './inactive-table'

export function InactiveTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { users = [], isLoading } = useStreamHotspotInactive(selectedDeviceId)

  return <InactiveTable data={users} isLoading={isLoading} />
}
