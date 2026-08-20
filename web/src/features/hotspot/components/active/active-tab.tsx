import { useStreamActiveSessions } from '../../api/use-hotspot-stream'
import { useDeviceStore } from '@/stores/device-store'
import { ActiveTable } from './active-table'

export function ActiveTab() {
  const { selectedDeviceId } = useDeviceStore()
  const { sessions = [], isLoading } = useStreamActiveSessions(selectedDeviceId)

  return <ActiveTable data={sessions} isLoading={isLoading} />
}
