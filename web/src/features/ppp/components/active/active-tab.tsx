import { useState } from 'react'
import { Activity, Radio } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useDeviceStore } from '@/stores/device-store'
import { usePPPActiveSessionsQuery } from '../../api/use-ppp-active'
import { useStreamPPPActiveSessions } from '../../api/use-ppp-stream'
import { ActiveTable } from './active-table'

export function ActiveTab() {
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const [liveMode, setLiveMode] = useState(true)

  // Query-based fallback
  const { data: polledSessions = [], isLoading: isQueryLoading } =
    usePPPActiveSessionsQuery(selectedDeviceId)

  // Stream-based live data
  const { sessions: streamedSessions, isLoading: isStreamLoading } =
    useStreamPPPActiveSessions(selectedDeviceId, liveMode)

  const sessions = liveMode ? (streamedSessions.length > 0 ? streamedSessions : polledSessions) : polledSessions
  const isLoading = liveMode ? isStreamLoading && polledSessions.length === 0 : isQueryLoading

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardHeader className="px-0 pt-0">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <CardTitle className="text-xl flex items-center gap-2">
              <Activity className="h-5 w-5 text-emerald-500" />
              Active PPPoE Sessions
              <Badge variant="outline" className="ml-2 font-mono">
                {sessions.length} Online
              </Badge>
            </CardTitle>
            <CardDescription>
              Real-time monitoring of connected subscriber CPEs, session uptimes, assigned IPs, and traffic stats.
            </CardDescription>
          </div>

          <div className="flex items-center gap-2 bg-muted/40 px-3 py-1.5 rounded-lg border">
            <Radio
              className={`h-4 w-4 ${liveMode ? 'text-emerald-500 animate-pulse' : 'text-muted-foreground'}`}
            />
            <Label htmlFor="live-toggle" className="text-xs cursor-pointer font-medium">
              Live Stream
            </Label>
            <Switch
              id="live-toggle"
              checked={liveMode}
              onCheckedChange={setLiveMode}
            />
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0">
        <ActiveTable data={sessions} isLoading={isLoading} />
      </CardContent>
    </Card>
  )
}
