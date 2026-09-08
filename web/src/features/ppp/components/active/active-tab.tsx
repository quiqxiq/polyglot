import { useMemo } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Route } from '@/routes/_authenticated/ppp/index'
import { useDeviceStore } from '@/stores/device-store'
import { usePPPActiveSessionsQuery } from '../../api/use-ppp-active'
import { usePPPSecretsQuery } from '../../api/use-ppp-secrets'
import { useStreamPPPActiveSessions, type EnrichedPPPActiveSession } from '../../api/use-ppp-stream'
import { ActiveTable } from './active-table'

export function ActiveTab() {
  const search = Route.useSearch()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)

  // Secrets list for profile enrichment fallback
  const { data: secrets = [] } = usePPPSecretsQuery(selectedDeviceId)
  const secretsProfileMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const s of secrets) {
      if (s.name && s.profile) {
        map.set(s.name, s.profile)
      }
    }
    return map
  }, [secrets])

  // Query-based fallback
  const { data: polledSessions = [], isLoading: isQueryLoading } =
    usePPPActiveSessionsQuery(selectedDeviceId)

  // Stream-based live data (always live)
  const { sessions: streamedSessions, isLoading: isStreamLoading } =
    useStreamPPPActiveSessions(selectedDeviceId, true)

  const rawSessions = streamedSessions.length > 0 ? streamedSessions : polledSessions
  const sessions = useMemo<EnrichedPPPActiveSession[]>(() => {
    return rawSessions.map((s) => {
      const resolvedProfile = s.profile && s.profile !== '' && s.profile !== 'default'
        ? s.profile
        : secretsProfileMap.get(s.name) || s.profile || 'default'
      const cloned = s.clone ? s.clone() : Object.assign(Object.create(Object.getPrototypeOf(s)), s)
      cloned.profile = resolvedProfile
      return cloned as EnrichedPPPActiveSession
    })
  }, [rawSessions, secretsProfileMap])

  const isLoading = isStreamLoading && polledSessions.length === 0 ? true : isQueryLoading

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardContent className="px-0">
        <ActiveTable
          data={sessions}
          isLoading={isLoading}
          defaultGlobalFilter={search.filter}
        />
      </CardContent>
    </Card>
  )
}

