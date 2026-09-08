import { Card, CardContent } from '@/components/ui/card'
import { useDeviceStore } from '@/stores/device-store'
import { Route } from '@/routes/_authenticated/ppp/index'
import { usePPPSecretsQuery } from '../../api/use-ppp-secrets'
import { SecretsTable } from './secrets-table'

export function SecretsTab() {
  const search = Route.useSearch()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: secrets = [], isLoading } = usePPPSecretsQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardContent className="px-0">
        <SecretsTable
          data={secrets}
          isLoading={isLoading}
          defaultGlobalFilter={search.filter}
        />
      </CardContent>
    </Card>
  )
}
