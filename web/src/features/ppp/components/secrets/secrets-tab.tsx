import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useDeviceStore } from '@/stores/device-store'
import { KeyRound } from 'lucide-react'
import { usePPPSecretsQuery } from '../../api/use-ppp-secrets'
import { SecretsTable } from './secrets-table'

export function SecretsTab() {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: secrets = [], isLoading } = usePPPSecretsQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardHeader className="px-0 pt-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-xl flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-primary" />
              PPPoE / PPP Secrets
            </CardTitle>
            <CardDescription>
              Manage subscriber credentials, assigned IP pools, bandwidth profiles, and account status.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0">
        <SecretsTable data={secrets} isLoading={isLoading} />
      </CardContent>
    </Card>
  )
}
