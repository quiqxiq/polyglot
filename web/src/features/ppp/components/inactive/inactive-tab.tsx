import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useDeviceStore } from '@/stores/device-store'
import { UserX } from 'lucide-react'
import { usePPPInactiveSecretsQuery } from '../../api/use-ppp-inactive'
import { InactiveTable } from './inactive-table'

export function InactiveTab() {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: inactive = [], isLoading } = usePPPInactiveSecretsQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardHeader className="px-0 pt-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-xl flex items-center gap-2">
              <UserX className="h-5 w-5 text-muted-foreground" />
              Offline PPPoE Subscribers
              <Badge variant="outline" className="ml-2 font-mono">
                {inactive.length} Offline
              </Badge>
            </CardTitle>
            <CardDescription>
              Subscribers who are registered in secrets but currently have no active session on the router.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0">
        <InactiveTable data={inactive} isLoading={isLoading} />
      </CardContent>
    </Card>
  )
}
