import { useEffect, useRef, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useDeviceStore } from '@/stores/device-store'
import { deviceClient } from '@/lib/api-client'
import { Activity, ArrowUpRight, Pause, Play, RefreshCw, Wifi, WifiOff } from 'lucide-react'
import { usePPP } from '../../context/ppp-context'

interface PingPacket {
  seq: number
  ttl?: number
  size?: number
  latencyMs: number
  status: string
  timestamp: Date
}

interface RouterSummaryStats {
  sent?: number
  received?: number
  packetLoss?: number
  minRttMs?: number
  avgRttMs?: number
  maxRttMs?: number
}

const MAX_HISTORY = 30

export function ActivePingDialog() {
  const { open, setOpen, currentActiveSession } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)

  const isOpen = open === 'active-ping'
  const [isRunning, setIsRunning] = useState(true)
  const [packets, setPackets] = useState<PingPacket[]>([])
  const [currentLatency, setCurrentLatency] = useState<number | null>(null)
  const [currentStatus, setCurrentStatus] = useState<string>('idle')
  const [routerStats, setRouterStats] = useState<RouterSummaryStats | null>(null)

  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const seqCounterRef = useRef(0)

  const targetIp = currentActiveSession?.address || ''
  const targetUser = currentActiveSession?.name || ''

  // Reset state on open/close
  useEffect(() => {
    if (isOpen) {
      setPackets([])
      setCurrentLatency(null)
      setCurrentStatus('connecting')
      setRouterStats(null)
      setIsRunning(true)
      seqCounterRef.current = 0
    }
  }, [isOpen, targetIp])

  // Stream Ping Effect
  useEffect(() => {
    if (!isOpen || !selectedDeviceId || !targetIp || !isRunning) return

    const controller = new AbortController()
    let active = true

    async function startStream() {
      try {
        setCurrentStatus('running')
        const stream = deviceClient.streamPing(
          { id: selectedDeviceId, address: targetIp },
          { signal: controller.signal }
        )

        for await (const frame of stream) {
          if (!active) break
          const lat = Number(frame.latencyMs)
          const status = frame.status || (lat >= 0 ? 'success' : 'timeout')
          const seq = frame.seq !== undefined && frame.seq !== 0 ? frame.seq : seqCounterRef.current + 1
          seqCounterRef.current = seq

          setCurrentLatency(lat >= 0 ? lat : null)
          setCurrentStatus(status)

          // Capture RouterOS native summary stats if present
          if (frame.sent > 0 || frame.received > 0 || frame.avgRttMs > 0 || frame.packetLoss > 0) {
            setRouterStats({
              sent: frame.sent,
              received: frame.received,
              packetLoss: frame.packetLoss,
              minRttMs: Number(frame.minRttMs),
              avgRttMs: Number(frame.avgRttMs),
              maxRttMs: Number(frame.maxRttMs),
            })
          }

          setPackets((prev) => {
            const newPacket: PingPacket = {
              seq,
              ttl: frame.ttl || undefined,
              size: frame.size || undefined,
              latencyMs: lat,
              status: status,
              timestamp: new Date(),
            }
            return [...prev.slice(-(MAX_HISTORY - 1)), newPacket]
          })
        }
      } catch (err: any) {
        if (err?.name !== 'AbortError' && active) {
          setCurrentStatus('error')
        }
      }
    }

    startStream()

    return () => {
      active = false
      controller.abort()
    }
  }, [isOpen, selectedDeviceId, targetIp, isRunning])

  // Render Sparkline on Canvas
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.max(100, Math.round(rect.width || 380))
    const h = Math.max(40, Math.round(rect.height || 70))

    if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
      canvas.width = w * dpr
      canvas.height = h * dpr
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)

    if (packets.length === 0) {
      ctx.fillStyle = '#94a3b8'
      ctx.font = '11px sans-serif'
      ctx.fillText('Waiting for ping replies...', 12, h / 2 + 4)
      return
    }

    const barW = w / MAX_HISTORY
    const validPings = packets.filter((p) => p.latencyMs >= 0).map((p) => p.latencyMs)
    const maxMs = Math.max(50, ...validPings)

    packets.forEach((p, i) => {
      const x = w - (packets.length - i) * barW
      if (p.latencyMs < 0 || p.status === 'timeout') {
        // Timeout bar (red line)
        ctx.fillStyle = '#ef4444'
        ctx.fillRect(x + 1, 0, Math.max(2, barW - 2), h)
      } else {
        const barH = Math.max(3, Math.min(h - 4, (p.latencyMs / maxMs) * (h - 4)))
        ctx.fillStyle = p.latencyMs > 80 ? '#f59e0b' : '#10b981'
        ctx.globalAlpha = 0.4 + 0.6 * ((i + 1) / packets.length)
        ctx.fillRect(x + 1, h - barH, Math.max(2, barW - 2), barH)
      }
    })
    ctx.globalAlpha = 1
  }, [packets])

  // Summary Metrics (Use native MikroTik routerStats if available, otherwise calculate from stream)
  const validPackets = packets.filter((p) => p.latencyMs >= 0)
  const sentCount = routerStats?.sent && routerStats.sent > 0 ? routerStats.sent : packets.length
  const receivedCount = routerStats?.received && routerStats.received > 0 ? routerStats.received : validPackets.length
  const lostCount = sentCount - receivedCount
  const lossPct = routerStats?.packetLoss !== undefined && routerStats.packetLoss >= 0
    ? routerStats.packetLoss
    : (sentCount > 0 ? Math.round((lostCount / sentCount) * 100) : 0)

  const latencies = validPackets.map((p) => p.latencyMs)
  const minLatency = routerStats?.minRttMs && routerStats.minRttMs > 0
    ? routerStats.minRttMs
    : (latencies.length ? Math.min(...latencies) : 0)
  const maxLatency = routerStats?.maxRttMs && routerStats.maxRttMs > 0
    ? routerStats.maxRttMs
    : (latencies.length ? Math.max(...latencies) : 0)
  const avgLatency = routerStats?.avgRttMs && routerStats.avgRttMs > 0
    ? routerStats.avgRttMs
    : (latencies.length ? Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length) : 0)

  return (
    <Dialog open={isOpen} onOpenChange={(v) => !v && setOpen(null)}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <div className="flex items-center justify-between pr-4">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-primary">
                <Activity className="h-4 w-4" />
              </div>
              <div>
                <DialogTitle className="text-base font-semibold">Live Ping Stream</DialogTitle>
                <DialogDescription className="text-xs font-mono">
                  {targetUser ? `${targetUser} • ` : ''}{targetIp}
                </DialogDescription>
              </div>
            </div>
            {currentLatency !== null ? (
              <Badge
                variant="outline"
                className={`font-mono text-xs px-2 py-0.5 ${
                  currentLatency > 80
                    ? 'border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400'
                    : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                }`}
              >
                <Wifi className="mr-1 h-3 w-3" />
                {currentLatency} ms
              </Badge>
            ) : (
              <Badge variant="outline" className="border-rose-500/30 bg-rose-500/10 text-rose-600 text-xs">
                <WifiOff className="mr-1 h-3 w-3" />
                {currentStatus === 'running' ? 'Request Timeout' : currentStatus}
              </Badge>
            )}
          </div>
        </DialogHeader>

        {/* Real-Time Sparkline Chart */}
        <div className="space-y-2">
          <div className="flex justify-between items-center text-xs text-muted-foreground">
            <span>Latency History (Last {MAX_HISTORY} pkts)</span>
            <span className="font-mono text-[11px]">{sentCount} sent / {receivedCount} recv</span>
          </div>
          <div className="h-20 w-full rounded-md border bg-muted/30 p-1.5 overflow-hidden">
            <canvas ref={canvasRef} className="h-full w-full block" />
          </div>
        </div>

        {/* Summary Stats Cards */}
        <div className="grid grid-cols-4 gap-2 text-center text-xs font-mono py-1">
          <div className="rounded-md border bg-card p-2">
            <div className="text-[10px] text-muted-foreground uppercase">Min RTT</div>
            <div className="font-semibold text-emerald-600 dark:text-emerald-400">{minLatency} ms</div>
          </div>
          <div className="rounded-md border bg-card p-2">
            <div className="text-[10px] text-muted-foreground uppercase">Avg RTT</div>
            <div className="font-semibold text-primary">{avgLatency} ms</div>
          </div>
          <div className="rounded-md border bg-card p-2">
            <div className="text-[10px] text-muted-foreground uppercase">Max RTT</div>
            <div className="font-semibold text-amber-600 dark:text-amber-400">{maxLatency} ms</div>
          </div>
          <div className="rounded-md border bg-card p-2">
            <div className="text-[10px] text-muted-foreground uppercase">Loss</div>
            <div className={`font-semibold ${lossPct > 0 ? 'text-rose-500' : 'text-emerald-600 dark:text-emerald-400'}`}>
              {lossPct}%
            </div>
          </div>
        </div>

        {/* Recent Packet Log Table */}
        <div className="rounded-md border overflow-hidden">
          <div className="max-h-40 overflow-y-auto font-mono text-xs divide-y divide-border/60">
            {packets.length === 0 ? (
              <div className="p-3 text-center text-muted-foreground">No packets yet</div>
            ) : (
              [...packets].reverse().slice(0, 10).map((p, idx) => (
                <div key={idx} className="flex items-center justify-between px-3 py-1.5 hover:bg-muted/40">
                  <div className="flex items-center gap-2">
                    <span className="text-muted-foreground">seq={p.seq}</span>
                    {p.size !== undefined && <span className="text-muted-foreground/80">{p.size}B</span>}
                    {p.ttl !== undefined && <span className="text-muted-foreground/80">ttl={p.ttl}</span>}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-muted-foreground text-[11px]">
                      {p.timestamp.toLocaleTimeString()}
                    </span>
                    {p.latencyMs >= 0 ? (
                      <span className="font-medium text-emerald-600 dark:text-emerald-400">
                        {p.latencyMs} ms
                      </span>
                    ) : (
                      <span className="font-medium text-rose-500">Timeout</span>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <DialogFooter className="flex items-center justify-between sm:justify-between w-full">
          <div className="flex items-center gap-1.5">
            <Button
              variant="outline"
              size="sm"
              className="h-8 text-xs"
              onClick={() => setIsRunning(!isRunning)}
            >
              {isRunning ? (
                <>
                  <Pause className="mr-1.5 h-3.5 w-3.5" /> Pause
                </>
              ) : (
                <>
                  <Play className="mr-1.5 h-3.5 w-3.5" /> Resume
                </>
              )}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 text-xs"
              onClick={() => {
                setPackets([])
                setCurrentLatency(null)
              }}
            >
              <RefreshCw className="mr-1.5 h-3.5 w-3.5" /> Clear
            </Button>
          </div>

          <div className="flex items-center gap-2">
            {targetIp && (
              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs"
                onClick={() => window.open(`http://${targetIp}`, '_blank')}
              >
                <ArrowUpRight className="mr-1.5 h-3.5 w-3.5" /> Open IP
              </Button>
            )}
            <Button size="sm" className="h-8 text-xs" onClick={() => setOpen(null)}>
              Close
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
