import { useState, useMemo, useCallback } from 'react'
import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Device } from '@/gen/v1/device_pb'
import { useDevicePingMetricsQuery } from '../../api/use-devices'
import { formatDateInput, formatTimeInput } from '../../lib/formatters'
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
	const [initialRange] = useState(() => {
		const end = new Date()
		return { start: new Date(end.getTime() - 60 * 60 * 1000), end }
	})
	const initialStart = initialRange.start
	const initialEnd = initialRange.end
	const [startDate, setStartDate] = useState<string>(() =>
		formatDateInput(initialStart)
	)
	const [startTime, setStartTime] = useState<string>(() =>
		formatTimeInput(initialStart)
	)
	const [endDate, setEndDate] = useState<string>(() => formatDateInput(initialEnd))
	const [endTime, setEndTime] = useState<string>(() => formatTimeInput(initialEnd))
	const [bucketInterval, setBucketInterval] = useState<string>('1m')
	const [selectedPreset, setSelectedPreset] = useState<TimePreset>('1h')

  const pingTarget = device?.pingTarget || device?.extra?.ping_target || '8.8.8.8'

	const handlePreset = useCallback((preset: TimePreset) => {
		setSelectedPreset(preset)
		if (preset === 'custom') return
		const cur = new Date()
		let past = new Date()
		let nextBucket = '1m'
		if (preset === '1h') {
			past = new Date(cur.getTime() - 60 * 60 * 1000)
			nextBucket = '1m'
		} else if (preset === '6h') {
			past = new Date(cur.getTime() - 6 * 60 * 60 * 1000)
			nextBucket = '1m'
		} else if (preset === '12h') {
			past = new Date(cur.getTime() - 12 * 60 * 60 * 1000)
			nextBucket = '5m'
		} else if (preset === 'today') {
			past = new Date(cur.getFullYear(), cur.getMonth(), cur.getDate(), 0, 0, 0)
			nextBucket = '5m'
		} else if (preset === '24h') {
			past = new Date(cur.getTime() - 24 * 60 * 60 * 1000)
			nextBucket = '5m'
		} else if (preset === '3d') {
			past = new Date(cur.getTime() - 3 * 24 * 60 * 60 * 1000)
			nextBucket = '5m'
		} else if (preset === '7d') {
			past = new Date(cur.getTime() - 7 * 24 * 60 * 60 * 1000)
			nextBucket = '5m'
		} else if (preset === '15d') {
			past = new Date(cur.getTime() - 15 * 24 * 60 * 60 * 1000)
			nextBucket = '1h'
		} else if (preset === '30d') {
			past = new Date(cur.getTime() - 30 * 24 * 60 * 60 * 1000)
			nextBucket = '1h'
		}
		setStartDate(formatDateInput(past))
		setStartTime(formatTimeInput(past))
		setEndDate(formatDateInput(cur))
		setEndTime(formatTimeInput(cur))
		setBucketInterval(nextBucket)
	}, [])

	// Calculate RFC3339 timestamps for historical query
	const startRFC = useMemo(() => {
		try {
			const sTime = startTime.includes(':') ? startTime : `${startTime}:00`
			const parts = sTime.split(':')
			const hh = (parts[0] || '00').padStart(2, '0')
			const mm = (parts[1] || '00').padStart(2, '0')
			const ss = (parts[2] || '00').padStart(2, '0')
			const d = new Date(`${startDate}T${hh}:${mm}:${ss}`)
			return isNaN(d.getTime()) ? undefined : d.toISOString()
		} catch {
			return undefined
		}
	}, [startDate, startTime])

	const endRFC = useMemo(() => {
		try {
			const eTime = endTime.includes(':') ? endTime : `${endTime}:00`
			const parts = eTime.split(':')
			const hh = (parts[0] || '23').padStart(2, '0')
			const mm = (parts[1] || '59').padStart(2, '0')
			const ss = (parts[2] || '59').padStart(2, '0')
			const d = new Date(`${endDate}T${hh}:${mm}:${ss}.999`)
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

	const points = metricsData?.points || []
	const summary: PingSummaryStats = {
		minRtt: metricsData?.minRtt || 0,
		avgRtt: metricsData?.avgRtt || 0,
		maxRtt: metricsData?.maxRtt || 0,
		packetLossPct: metricsData?.packetLossPct || 0,
		totalSamples: Number(metricsData?.totalSamples || 0),
	}

  if (!device) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-5xl lg:max-w-6xl w-[95vw] max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden'>
        {/* Header */}
        <PingAnalyticsHeader
          deviceName={device.name}
          host={device.host}
          pingTarget={pingTarget}
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
          <PingAnalyticsKpiGrid summary={summary} />

          {/* Charts & Log Table Tabs */}
          <Tabs defaultValue='chart' className='w-full'>
            <div className='flex items-center justify-between pb-2'>
              <TabsList className='h-8'>
                <TabsTrigger value='chart' className='text-xs h-7 px-3'>
                  Grafik Latensi
                </TabsTrigger>
                <TabsTrigger value='table' className='text-xs h-7 px-3'>
                  Riwayat Paket ({points.length})
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value='chart' className='mt-0'>
              <PingAnalyticsChart
                points={points}
                avgRtt={summary.avgRtt}
                isLoading={isLoading}
              />
            </TabsContent>

            <TabsContent value='table' className='mt-0'>
              <PingAnalyticsTable points={points} />
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  )
}
