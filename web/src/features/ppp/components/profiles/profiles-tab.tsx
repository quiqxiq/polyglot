import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useDeviceStore } from '@/stores/device-store'
import { Shield } from 'lucide-react'
import { usePPPProfilesQuery } from '../../api/use-ppp-profiles'
import { ProfilesTable } from './profiles-table'

export function ProfilesTab() {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: profiles = [], isLoading } = usePPPProfilesQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardHeader className="px-0 pt-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-xl flex items-center gap-2">
              <Shield className="h-5 w-5 text-primary" />
              PPP Bandwidth & IP Profiles
              <Badge variant="outline" className="ml-2 font-mono">
                {profiles.length} Profiles
              </Badge>
            </CardTitle>
            <CardDescription>
              Define bandwidth shaping (Rx/Tx rate limits), assigned address pools, DNS push settings, and session concurrency rules.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0">
        <ProfilesTable data={profiles} isLoading={isLoading} />
      </CardContent>
    </Card>
  )
}
