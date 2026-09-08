import { Card, CardContent } from '@/components/ui/card'
import { useDeviceStore } from '@/stores/device-store'
import { usePPPProfilesQuery } from '../../api/use-ppp-profiles'
import { ProfilesTable } from './profiles-table'

export function ProfilesTab() {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: profiles = [], isLoading } = usePPPProfilesQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardContent className="px-0">
        <ProfilesTable data={profiles} isLoading={isLoading} />
      </CardContent>
    </Card>
  )
}
