import { Card, CardContent } from '@/components/ui/card'
import { useDeviceStore } from '@/stores/device-store'
import { Route } from '@/routes/_authenticated/ppp/index'
import { usePPPInactiveSecretsQuery } from '../../api/use-ppp-inactive'
import { InactiveTable } from './inactive-table'

export function InactiveTab() {
  const search = Route.useSearch()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: inactive = [], isLoading } = usePPPInactiveSecretsQuery(selectedDeviceId)

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardContent className="px-0">
        <InactiveTable
          data={inactive}
          isLoading={isLoading}
          defaultGlobalFilter={search.filter}
        />
      </CardContent>
    </Card>
  )
}

