import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Device, PingMetricPointData } from '@/gen/v1/device_pb'
import { deviceClient } from '@/lib/api-client'
import { useDevicePingMetricsQuery } from '../api/use-devices'
import {
  Activity,
  AlertTriangle,
  ArrowDown,
  ArrowUp,
  Calendar,
  CheckCircle2,
  Clock,
  Layers,
  Loader2,
  Radio,
  RefreshCw,
  TrendingDown,
  Wifi,
} from 'lucide-react'

interface DevicePingAnalyticsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  device: Device
}

function formatDateInput(d: Date): string {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

function formatTimeInput(d: Date): string {
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

export function DevicePingAnalyticsDialog({
  open,
  onOpenChange,
  device,
}: DevicePingAnalyticsDialogProps) {
  const [startDate, setStartDate] = useState<string>(() => formatDateInput(new Date(Date.now() - 60 * 60 * 1000)))
  const [startTime, setStartTime] = useState<string>(() => formatTimeInput(new Date(Date.now() - 60 * 60 * 1000)))
  const [endDate, setEndDate] = useState<string>(() => formatDateInput(new Date()))
  const [endTime, setEndTime] = useState<string>(() => formatTimeInput(new Date()))
  const [bucketInterval, setBucketInterval] = useState<string>('raw')
  const [selectedPreset, setSelectedPreset] = useState<string>('1h')
  const [isStreaming, setIsStreaming] = useState<boolean>(false)
  const [livePoints, setLivePoints] = useState<PingMetricPointData[]>([])

  const pingTarget = device.pingTarget || device.extra?.ping_target || '8.8.8.8'

  const handlePreset = useCallback((preset: string) => {
    setSelectedPreset(preset)
    const cur = new Date()
    let past = new Date()
    if (preset === '1h') {
      past = new Date(cur.getTime() - 60 * 60 * 1000)
    } else if (preset === '6h') {
      past = new Date(cur.getTime() - 6 * 60 * 60 * 1000)
    } else if (preset === 'today') {
      past = new Date(cur.getFullYear(), cur.getMonth(), cur.getDate(), 0, 0, 0)
    } else if (preset === '24h') {
      past = new Date(cur.getTime() - 24 * 60 * 60 * 1000)
    } else if (preset === '7d') {
      past = new Date(cur.getTime() - 7 * 24 * 60 * 60 * 1000)
      setBucketInterval('5m')
    }
    setStartDate(formatDateInput(past))
    setStartTime(formatTimeInput(past))
    setEndDate(formatDateInput(cur))
    setEndTime(formatTimeInput(cur))
    setLivePoints([])
  }, [])

  useEffect(() => {
    if (open) {
      handlePreset('1h')
    } else {
      setLivePoints([])
    }
  }, [open, handlePreset])

  // Calculate RFC3339 timestamps for historical initial load
  const startRFC = useMemo(() => {
    try {
      const d = new Date(`${startDate}T${startTime}:00`)
      return isNaN(d.getTime()) ? undefined : d.toISOString()
    } catch {
      return undefined
    }
  }, [startDate, startTime])

  const endRFC = useMemo(() => {
    try {
      const d = new Date(`${endDate}T${endTime}:59.999`)
      return isNaN(d.getTime()) ? undefined : d.toISOString()
    } catch {
      return undefined
    }
  }, [endDate, endTime])

  const { data: metricsData, isLoading, refetch, isFetching } = useDevicePingMetricsQuery(
    {
      deviceId: device.id,
      startTime: startRFC,
      endTime: endRFC,
      bucketInterval: bucketInterval === 'raw' ? '' : bucketInterval,
    },
    open
  )

  // Native ConnectRPC Streaming Ping frames (No Polling)
  const isStreamingActiveRef = useRef(false)
  useEffect(() => {
    if (!open || !device.id || !device.enabled || !pingTarget) return
    const controller = new AbortController()
    isStreamingActiveRef.current = true
    setIsStreaming(true)

    async function runLiveStream() {
      try {
        const stream = deviceClient.streamPing(
          { id: device.id, address: pingTarget },
          { signal: controller.signal }
        )
        for await (const frame of stream) {
          if (!isStreamingActiveRef.current) break
          const lat = Number(frame.latencyMs)
          const newPt = new PingMetricPointData({
            timestamp: new Date().toISOString(),
            target: frame.address || pingTarget,
            seq: Number(frame.seq) || 0,
            size: Number(frame.size) || 56,
            ttl: Number(frame.ttl) || 116,
            rttMs: lat,
            status: frame.status || 'connected',
            sent: Number(frame.sent) || 1,
            received: Number(frame.received) || 1,
            packetLoss: Number(frame.packetLoss) || 0,
            minRttMs: Number(frame.minRttMs) || lat,
            avgRttMs: Number(frame.avgRttMs) || lat,
            maxRttMs: Number(frame.maxRttMs) || lat,
          })

          setLivePoints((prev) => {
            const next = [...prev, newPt]
            return next.slice(-200) // retain last 200 live points in memory
          })
        }
      } catch (err: any) {
        const isAborted = controller.signal.aborted || !isStreamingActiveRef.current || err?.name === 'AbortError'
        if (!isAborted && isStreamingActiveRef.current) {
          setTimeout(() => {
            if (isStreamingActiveRef.current && !controller.signal.aborted) {
              runLiveStream()
            }
          }, 3000)
        }
      }
    }

    runLiveStream()

    return () => {
      isStreamingActiveRef.current = false
      setIsStreaming(false)
      controller.abort()
    }
  }, [open, device.id, device.enabled, pingTarget])

  // Merge historical points with live stream points
  const mergedPoints = useMemo(() => {
    const hist = metricsData?.points || []
    if (livePoints.length === 0) return hist
    // Combine and deduplicate if timestamps overlap
    const seen = new Set<string>()
    const res: PingMetricPointData[] = []
    for (const p of hist) {
      const key = `${p.timestamp}_${p.seq}`
      if (!seen.has(key)) {
        seen.add(key)
        res.push(p)
      }
    }
    for (const p of livePoints) {
      const key = `${p.timestamp}_${p.seq}`
      if (!seen.has(key)) {
        seen.add(key)
        res.push(p)
      }
    }
    return res
  }, [metricsData?.points, livePoints])

  const timescaledbAvailable = metricsData?.timescaledbAvailable ?? true

  // Calculated Live Summary Statistics
  const dynamicSummary = useMemo(() => {
    if (!mergedPoints.length) {
      return {
        minRtt: metricsData?.minRtt || 0,
        avgRtt: metricsData?.avgRtt || 0,
        maxRtt: metricsData?.maxRtt || 0,
        packetLossPct: metricsData?.packetLossPct || 0,
        totalSamples: Number(metricsData?.totalSamples || 0) + livePoints.length,
      }
    }
    const valid = mergedPoints.filter((p) => p.rttMs > 0)
    const rtts = valid.map((p) => p.rttMs)
    const min = rtts.length ? Math.min(...rtts) : 0
    const max = rtts.length ? Math.max(...rtts) : 0
    const sum = rtts.reduce((a, b) => a + b, 0)
    const avg = rtts.length ? sum / rtts.length : 0
    const lossSum = mergedPoints.reduce((a, b) => a + b.packetLoss, 0)
    const lossPct = mergedPoints.length ? (lossSum / mergedPoints.length) : 0

    return {
      minRtt: min,
      avgRtt: avg,
      maxRtt: max,
      packetLossPct: lossPct,
      totalSamples: Math.max(mergedPoints.length, Number(metricsData?.totalSamples || 0)),
    }
  }, [mergedPoints, metricsData, livePoints.length])

  // Simple SVG Line Chart generator
  const chartPoints = useMemo(() => {
    if (!mergedPoints || mergedPoints.length === 0) return []
    return mergedPoints.map((p, idx) => ({
      x: idx,
      y: p.rttMs,
      loss: p.packetLoss,
      time: p.timestamp,
      status: p.status,
    }))
  }, [mergedPoints])

  const maxVal = useMemo(() => {
    if (!chartPoints.length) return 100
    const m = Math.max(...chartPoints.map((p) => p.y))
    return m > 0 ? Math.ceil(m * 1.2) : 50
  }, [chartPoints])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden">
        <DialogHeader className="p-6 pb-4 border-b">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="rounded-lg bg-primary/10 p-2 text-primary">
                <Activity className="h-5 w-5" />
              </div>
              <div>
                <DialogTitle className="text-base font-semibold flex items-center gap-2">
                  Analisis Ping Streaming — {device.name}
                  {isStreaming && (
                    <Badge variant="outline" className="text-[10px] bg-emerald-500/10 text-emerald-600 border-emerald-300 gap-1 animate-pulse">
                      <Radio className="h-2.5 w-2.5" />
                      Live Stream
                    </Badge>
                  )}
                </DialogTitle>
                <DialogDescription className="text-xs text-muted-foreground flex items-center gap-2 mt-0.5">
                  <span>Target: <code className="font-mono text-primary font-semibold">{pingTarget}</code></span>
                  <span>•</span>
                  <span>Host: <code className="font-mono">{device.host}</code></span>
                </DialogDescription>
              </div>
            </div>

            <div className="flex items-center gap-2">
              {timescaledbAvailable ? (
                <Badge variant="outline" className="bg-emerald-500/10 text-emerald-600 border-emerald-200 text-xs gap-1">
                  <CheckCircle2 className="h-3 w-3" />
                  TimescaleDB Aktif
                </Badge>
              ) : (
                <Badge variant="destructive" className="text-xs gap-1">
                  <AlertTriangle className="h-3 w-3" />
                  TimescaleDB Tidak Aktif
                </Badge>
              )}
              <Button
                variant="outline"
                size="icon"
                className="h-8 w-8"
                onClick={() => refetch()}
                disabled={isFetching}
                title="Muat Ulang Data Historis"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${isFetching ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </div>
        </DialogHeader>

        {/* Filter Controls */}
        <div className="p-4 bg-muted/20 border-b flex flex-wrap items-center gap-3">
          {/* Quick Presets */}
          <div className="flex items-center gap-1 bg-background border rounded-md p-0.5">
            {[
              { id: '1h', label: '1 Jam' },
              { id: '6h', label: '6 Jam' },
              { id: 'today', label: 'Hari Ini' },
              { id: '24h', label: '24 Jam' },
              { id: '7d', label: '7 Hari' },
            ].map((p) => (
              <Button
                key={p.id}
                variant={selectedPreset === p.id ? 'secondary' : 'ghost'}
                size="sm"
                className="h-7 text-xs px-2.5"
                onClick={() => handlePreset(p.id)}
              >
                {p.label}
              </Button>
            ))}
          </div>

          {/* Date & Time Range Pickers */}
          <div className="flex items-center gap-2 bg-background border rounded-md px-3 py-1 text-xs">
            <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-muted-foreground">Dari:</span>
            <input
              type="date"
              value={startDate}
              onChange={(e) => {
                setStartDate(e.target.value)
                setSelectedPreset('custom')
              }}
              className="bg-transparent border-0 text-xs font-mono focus:outline-none"
            />
            <input
              type="time"
              value={startTime}
              onChange={(e) => {
                setStartTime(e.target.value)
                setSelectedPreset('custom')
              }}
              className="bg-transparent border-0 text-xs font-mono focus:outline-none"
            />

            <span className="text-muted-foreground mx-1">s/d</span>

            <input
              type="date"
              value={endDate}
              onChange={(e) => {
                setEndDate(e.target.value)
                setSelectedPreset('custom')
              }}
              className="bg-transparent border-0 text-xs font-mono focus:outline-none"
            />
            <input
              type="time"
              value={endTime}
              onChange={(e) => {
                setEndTime(e.target.value)
                setSelectedPreset('custom')
              }}
              className="bg-transparent border-0 text-xs font-mono focus:outline-none"
            />
          </div>

          {/* Bucket interval */}
          <div className="flex items-center gap-1.5 ml-auto">
            <Layers className="h-3.5 w-3.5 text-muted-foreground" />
            <Select value={bucketInterval} onValueChange={setBucketInterval}>
              <SelectTrigger className="h-8 text-xs w-28">
                <SelectValue placeholder="Resolusi" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="raw">Raw (Detik)</SelectItem>
                <SelectItem value="1m">1 Menit</SelectItem>
                <SelectItem value="5m">5 Menit</SelectItem>
                <SelectItem value="1h">1 Jam</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {!timescaledbAvailable && (
            <div className="rounded-lg bg-amber-500/10 p-3.5 border border-amber-500/20 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 shrink-0" />
              <div>
                <strong>Peringatan TimescaleDB:</strong> Ekstensi TimescaleDB belum diaktifkan di PostgreSQL. Fitur penyimpanan telemetri memerlukan TimescaleDB aktif.
              </div>
            </div>
          )}

          {/* KPI Cards */}
          <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
            <div className="rounded-lg border p-3 bg-card shadow-xs">
              <span className="text-[11px] text-muted-foreground">Min Latency</span>
              <div className="text-lg font-bold font-mono text-emerald-600 dark:text-emerald-400 flex items-center gap-1 mt-0.5">
                <ArrowDown className="h-3.5 w-3.5" />
                {`${dynamicSummary.minRtt.toFixed(1)} ms`}
              </div>
            </div>

            <div className="rounded-lg border p-3 bg-card shadow-xs">
              <span className="text-[11px] text-muted-foreground">Avg Latency</span>
              <div className="text-lg font-bold font-mono text-primary flex items-center gap-1 mt-0.5">
                <Wifi className="h-3.5 w-3.5" />
                {`${dynamicSummary.avgRtt.toFixed(1)} ms`}
              </div>
            </div>

            <div className="rounded-lg border p-3 bg-card shadow-xs">
              <span className="text-[11px] text-muted-foreground">Max Latency</span>
              <div className="text-lg font-bold font-mono text-amber-600 dark:text-amber-400 flex items-center gap-1 mt-0.5">
                <ArrowUp className="h-3.5 w-3.5" />
                {`${dynamicSummary.maxRtt.toFixed(1)} ms`}
              </div>
            </div>

            <div className="rounded-lg border p-3 bg-card shadow-xs">
              <span className="text-[11px] text-muted-foreground">Packet Loss</span>
              <div className={`text-lg font-bold font-mono flex items-center gap-1 mt-0.5 ${dynamicSummary.packetLossPct > 0 ? 'text-destructive' : 'text-emerald-600'}`}>
                <TrendingDown className="h-3.5 w-3.5" />
                {`${dynamicSummary.packetLossPct.toFixed(1)}%`}
              </div>
            </div>

            <div className="rounded-lg border p-3 bg-card shadow-xs col-span-2 sm:col-span-1">
              <span className="text-[11px] text-muted-foreground">Total Samples</span>
              <div className="text-lg font-bold font-mono text-foreground flex items-center gap-1 mt-0.5">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                {dynamicSummary.totalSamples.toLocaleString()}
              </div>
            </div>
          </div>

          {/* Charts & Table Tabs */}
          <Tabs defaultValue="chart" className="w-full">
            <div className="flex items-center justify-between pb-2">
              <TabsList className="h-8">
                <TabsTrigger value="chart" className="text-xs h-7">Grafik Latensi</TabsTrigger>
                <TabsTrigger value="table" className="text-xs h-7">Riwayat Paket ({mergedPoints.length})</TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="chart" className="mt-0">
              {isLoading && mergedPoints.length === 0 ? (
                <div className="h-64 flex items-center justify-center border rounded-lg">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : chartPoints.length === 0 ? (
                <div className="h-64 flex flex-col items-center justify-center border rounded-lg text-muted-foreground text-xs gap-1.5">
                  <Activity className="h-8 w-8 text-muted-foreground/40" />
                  <span>Tidak ada data metrik ping pada rentang waktu ini.</span>
                  <span className="text-[11px]">Pastikan fitur ping metrics pada router telah diaktifkan di Pengaturan.</span>
                </div>
              ) : (
                <div className="border rounded-lg p-4 bg-card space-y-4">
                  {/* SVG Line Chart */}
                  <div className="relative h-56 w-full">
                    <svg className="w-full h-full overflow-visible" viewBox={`0 0 ${Math.max(chartPoints.length, 1)} 100`} preserveAspectRatio="none">
                      {/* Grid Lines */}
                      <line x1="0" y1="25" x2={chartPoints.length} y2="25" stroke="currentColor" strokeOpacity="0.08" />
                      <line x1="0" y1="50" x2={chartPoints.length} y2="50" stroke="currentColor" strokeOpacity="0.08" />
                      <line x1="0" y1="75" x2={chartPoints.length} y2="75" stroke="currentColor" strokeOpacity="0.08" />

                      {/* Area under line */}
                      <polygon
                        points={`0,100 ${chartPoints
                          .map((p) => `${p.x},${100 - (p.y / maxVal) * 100}`)
                          .join(' ')} ${chartPoints.length - 1},100`}
                        fill="currentColor"
                        className="text-primary/10"
                      />

                      {/* Latency Line */}
                      <polyline
                        fill="none"
                        stroke="currentColor"
                        className="text-primary"
                        strokeWidth="1.5"
                        points={chartPoints
                          .map((p) => `${p.x},${100 - (p.y / maxVal) * 100}`)
                          .join(' ')}
                      />

                      {/* Average Reference Line */}
                      {dynamicSummary.avgRtt > 0 && (
                        <line
                          x1="0"
                          y1={100 - (dynamicSummary.avgRtt / maxVal) * 100}
                          x2={chartPoints.length}
                          y2={100 - (dynamicSummary.avgRtt / maxVal) * 100}
                          stroke="#10b981"
                          strokeDasharray="4 4"
                          strokeWidth="1"
                        />
                      )}
                    </svg>

                    <div className="absolute top-1 left-2 text-[10px] font-mono text-muted-foreground">
                      Max: {maxVal} ms
                    </div>
                    {dynamicSummary.avgRtt > 0 && (
                      <div className="absolute right-2 text-[10px] font-mono text-emerald-600 dark:text-emerald-400" style={{ top: `${Math.max(10, Math.min(85, 100 - (dynamicSummary.avgRtt / maxVal) * 100))}%` }}>
                        Avg: {dynamicSummary.avgRtt.toFixed(1)} ms
                      </div>
                    )}
                  </div>

                  <div className="flex items-center justify-between text-[11px] text-muted-foreground font-mono px-1">
                    <span>{mergedPoints[0]?.timestamp ? new Date(mergedPoints[0].timestamp).toLocaleTimeString() : ''}</span>
                    <span>{mergedPoints[mergedPoints.length - 1]?.timestamp ? new Date(mergedPoints[mergedPoints.length - 1].timestamp).toLocaleTimeString() : ''}</span>
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="table" className="mt-0">
              <div className="border rounded-lg overflow-hidden">
                <div className="max-h-64 overflow-y-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-muted/40 text-muted-foreground sticky top-0 border-b">
                      <tr>
                        <th className="p-2.5 font-medium">Waktu</th>
                        <th className="p-2.5 font-medium">SEQ</th>
                        <th className="p-2.5 font-medium">Host</th>
                        <th className="p-2.5 font-medium">Size</th>
                        <th className="p-2.5 font-medium">TTL</th>
                        <th className="p-2.5 font-medium">Latency</th>
                        <th className="p-2.5 font-medium">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y font-mono text-[11px]">
                      {mergedPoints.length === 0 ? (
                        <tr>
                          <td colSpan={7} className="p-4 text-center text-muted-foreground font-sans">
                            Tidak ada log paket pada rentang ini.
                          </td>
                        </tr>
                      ) : (
                        mergedPoints.slice(-100).reverse().map((p, i) => (
                          <tr key={i} className="hover:bg-muted/30">
                            <td className="p-2 text-muted-foreground">{p.timestamp ? new Date(p.timestamp).toLocaleTimeString() : '-'}</td>
                            <td className="p-2">{p.seq}</td>
                            <td className="p-2 text-primary">{p.target}</td>
                            <td className="p-2">{p.size}B</td>
                            <td className="p-2">{p.ttl}</td>
                            <td className="p-2 font-semibold text-foreground">{p.rttMs.toFixed(1)} ms</td>
                            <td className="p-2">
                              <Badge
                                variant="outline"
                                className={`text-[10px] px-1.5 py-0 ${
                                  p.status === 'connected' || p.status === 'ok'
                                    ? 'bg-emerald-500/10 text-emerald-600 border-emerald-200'
                                    : 'bg-destructive/10 text-destructive border-destructive/20'
                                }`}
                              >
                                {p.status}
                              </Badge>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  )
}
