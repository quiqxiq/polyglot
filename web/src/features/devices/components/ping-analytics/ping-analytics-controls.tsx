import { useMemo } from 'react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DateRangePicker, type DateRange, type DateRangePreset } from '@/components/ui/date-range-picker'
import { Layers } from 'lucide-react'
import type { TimePreset } from '../../types'
import { formatDateInput, formatTimeInput } from '../../lib/formatters'

interface PingAnalyticsControlsProps {
  selectedPreset: TimePreset
  onPresetChange: (preset: TimePreset) => void
  startDate: string
  startTime: string
  endDate: string
  endTime: string
  onStartDateChange: (val: string) => void
  onStartTimeChange: (val: string) => void
  onEndDateChange: (val: string) => void
  onEndTimeChange: (val: string) => void
  bucketInterval: string
  onBucketIntervalChange: (val: string) => void
}

const PING_TIME_PRESETS: { id: TimePreset; label: string }[] = [
  { id: '1h', label: '1 Jam' },
  { id: '6h', label: '6 Jam' },
  { id: '12h', label: '12 Jam' },
  { id: '24h', label: '24 Jam' },
  { id: 'today', label: 'Hari Ini' },
  { id: '3d', label: '3 Hari' },
  { id: '7d', label: '7 Hari' },
  { id: '15d', label: '15 Hari' },
]

const DATE_PICKER_PRESETS: DateRangePreset[] = [
  ...PING_TIME_PRESETS.map((p) => ({ name: p.id, label: p.label })),
  { name: 'yesterday', label: 'Kemarin' },
  { name: '30d', label: '30 Hari' },
]

export function PingAnalyticsControls({
  selectedPreset,
  onPresetChange,
  startDate,
  startTime,
  endDate,
  endTime,
  onStartDateChange,
  onStartTimeChange,
  onEndDateChange,
  onEndTimeChange,
  bucketInterval,
  onBucketIntervalChange,
}: PingAnalyticsControlsProps) {
  // Derive Date objects including time for DateRangePicker
  const dateFrom = useMemo(() => {
    if (!startDate) return new Date()
    const sTime = startTime ? (startTime.includes(':') ? startTime : `${startTime}:00`) : '00:00:00'
    const d = new Date(`${startDate}T${sTime}`)
    return isNaN(d.getTime()) ? new Date() : d
  }, [startDate, startTime])

  const dateTo = useMemo(() => {
    if (!endDate) return new Date()
    const eTime = endTime ? (endTime.includes(':') ? endTime : `${endTime}:00`) : '23:59:59'
    const d = new Date(`${endDate}T${eTime}`)
    return isNaN(d.getTime()) ? new Date() : d
  }, [endDate, endTime])

  const handleRangeUpdate = ({
    range,
    preset,
  }: {
    range: DateRange
    rangeCompare?: DateRange
    preset?: string
  }) => {
    if (range.from) {
      onStartDateChange(formatDateInput(range.from))
      onStartTimeChange(formatTimeInput(range.from))
    }
    const end = range.to || range.from
    if (end) {
      onEndDateChange(formatDateInput(end))
      onEndTimeChange(formatTimeInput(end))
    }
    if (preset) {
      onPresetChange(preset as TimePreset)
    } else {
      onPresetChange('custom')
    }
  }

  return (
    <div className='p-4 bg-muted/20 border-b flex flex-wrap items-center justify-between gap-3'>
      {/* Left section: Quick Presets & Unified Date-Time Range Picker */}
      <div className='flex flex-wrap items-center gap-2'>
        {/* Quick Presets */}
        <div className='flex items-center gap-1 bg-background border rounded-md p-0.5 shadow-2xs overflow-x-auto max-w-full'>
          {PING_TIME_PRESETS.map((p) => (
            <Button
              key={p.id}
              variant={selectedPreset === p.id ? 'secondary' : 'ghost'}
              size='sm'
              className='h-7 text-xs px-2.5 font-medium whitespace-nowrap'
              onClick={() => onPresetChange(p.id)}
            >
              {p.label}
            </Button>
          ))}
        </div>

        {/* Custom DateRangePicker with Time Selection & Presets */}
        <DateRangePicker
          initialDateFrom={dateFrom}
          initialDateTo={dateTo}
          selectedPreset={selectedPreset === 'custom' ? undefined : selectedPreset}
          presets={DATE_PICKER_PRESETS}
          showTimePicker={true}
          showCompare={false}
          locale='id-ID'
          align='start'
          size='sm'
          className='h-8 bg-background'
          onUpdate={handleRangeUpdate}
        />
      </div>

      {/* Right section: Bucket Interval Resolution */}
      <div className='flex items-center gap-1.5'>
        <Layers className='h-3.5 w-3.5 text-muted-foreground' />
        <Select value={bucketInterval} onValueChange={onBucketIntervalChange}>
          <SelectTrigger className='h-8 text-xs w-28 bg-background shadow-2xs'>
            <SelectValue placeholder='Resolusi' />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value='raw'>Raw (Detik)</SelectItem>
            <SelectItem value='1m'>1 Menit</SelectItem>
            <SelectItem value='5m'>5 Menit</SelectItem>
            <SelectItem value='1h'>1 Jam</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
