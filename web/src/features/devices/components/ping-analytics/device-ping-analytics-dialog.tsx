import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PingMetricPointData, type Device } from '@/gen/v1/device_pb'
import { deviceClient } from '@/lib/api-client'
import { useDevicePingMetricsQuery } from '../../api/use-devices'
import { formatDateInput, formatTimeInput } from '../../lib/formatters'
import { isStreamAbortedError } from '../../lib/stream-utils'
import type { PingSummaryStats, TimePreset } from '../../types'
import { PingAnalyticsHeader } from './ping-analytics-header'
import { PingAnalyticsControls } from './ping-analytics-controls'
import { PingAnalyticsKpiGrid } from './ping-analytics-kpi-grid'
import { PingAnalyticsChart } from './ping-analytics-chart'
import { PingAnalyticsTable } from './ping-analytics-table'

interface DevicePingAnalyticsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  device: Device | null
}

export function DevicePingAnalyticsDialog({
  open,
  onOpenChange,
  device,
}: DevicePingAnalyticsDialogProps) {
  const [startDate, setStartDate] = useState<string>(() =>
    formatDateInput(new Date(Date.now() - 60 * 60 * 1000))
  )
  const [startTime, setStartTime] = useState<string>(() =>
    formatTimeInput(new Date(Date.now() - 60 * 60 * 1000))
  )
  const [endDate, setEndDate] = useState<string>(() => formatDateInput(new Date()))
  const [endTime, setEndTime] = useState<string>(() => formatTimeInput(new Date()))
  const [bucketInterval, setBucketInterval] = useState<string>('raw')
  const [selectedPreset, setSelectedPreset] = useState<TimePreset>('1h')
  const [isStreaming, setIsStreaming] = useState<boolean>(false)
  const [livePoints, setLivePoints] = useState<PingMetricPointData[]>([])

  const pingTarget = device?.pingTarget || device?.extra?.ping_target || '8.8.8.8'

  const handlePreset = useCallback((preset: TimePreset) => {
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

  const {
    data: metricsData,
    isLoading,
    refetch,
    isFetching,
  } = useDevicePingMetricsQuery(
    {
      deviceId: device?.id || '',
      startTime: startRFC,
      endTime: endRFC,
      bucketInterval: bucketInterval === 'raw' ? '' : bucketInterval,
    },
    open && Boolean(device?.id)
  )

  // Native ConnectRPC Streaming Ping frames
  const isStreamingActiveRef = useRef(false)
  useEffect(() => {
    if (!open || !device?.id || !device.enabled || !pingTarget) {
      return
    }
    const controller = new AbortController()
    isStreamingActiveRef.current = true

    async function runLiveStream() {
      if (!device) return
      setIsStreaming(true)
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
            return next.slice(-200)
          })
        }
      } catch (err: unknown) {
        if (!isStreamAbortedError(err, controller.signal, isStreamingActiveRef.current) && isStreamingActiveRef.current) {
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
  }, [open, device?.id, device?.enabled, pingTarget, device])

  // Merge historical points with live stream points
  const mergedPoints = useMemo(() => {
    const hist = metricsData?.points || []
    if (livePoints.length === 0) return hist

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
  const dynamicSummary: PingSummaryStats = useMemo(() => {
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
    const lossPct = mergedPoints.length ? lossSum / mergedPoints.length : 0

    return {
      minRtt: min,
      avgRtt: avg,
      maxRtt: max,
      packetLossPct: lossPct,
      totalSamples: Math.max(
        mergedPoints.length,
        Number(metricsData?.totalSamples || 0)
      ),
    }
  }, [mergedPoints, metricsData, livePoints.length])

  if (!device) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-4xl max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden'>
        {/* Header */}
        <PingAnalyticsHeader
          deviceName={device.name}
          host={device.host}
          pingTarget={pingTarget}
          isStreaming={isStreaming}
          timescaledbAvailable={timescaledbAvailable}
          isFetching={isFetching}
          onRefresh={() => refetch()}
        />

        {/* Filter Toolbar Controls */}
        <PingAnalyticsControls
          selectedPreset={selectedPreset}
          onPresetChange={handlePreset}
          startDate={startDate}
          startTime={startTime}
          endDate={endDate}
          endTime={endTime}
          onStartDateChange={setStartDate}
          onStartTimeChange={setStartTime}
          onEndDateChange={setEndDate}
          onEndTimeChange={setEndTime}
          bucketInterval={bucketInterval}
          onBucketIntervalChange={setBucketInterval}
        />

        {/* Main Content Area */}
        <div className='flex-1 overflow-y-auto p-6 space-y-5'>
          {/* KPI Summary Cards */}
          <PingAnalyticsKpiGrid summary={dynamicSummary} />

          {/* Charts & Log Table Tabs */}
          <Tabs defaultValue='chart' className='w-full'>
            <div className='flex items-center justify-between pb-2'>
              <TabsList className='h-8'>
                <TabsTrigger value='chart' className='text-xs h-7 px-3'>
                  Grafik Latensi
                </TabsTrigger>
                <TabsTrigger value='table' className='text-xs h-7 px-3'>
                  Riwayat Paket ({mergedPoints.length})
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value='chart' className='mt-0'>
              <PingAnalyticsChart
                points={mergedPoints}
                avgRtt={dynamicSummary.avgRtt}
                isLoading={isLoading}
              />
            </TabsContent>

            <TabsContent value='table' className='mt-0'>
              <PingAnalyticsTable points={mergedPoints} />
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  )
}

