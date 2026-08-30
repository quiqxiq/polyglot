import { useState } from 'react'
import { format, parseISO } from 'date-fns'
import type { DateRange } from 'react-day-picker'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { CalendarIcon, Clock, Layers } from 'lucide-react'
import type { TimePreset } from '../../types'
import { formatDateInput } from '../../lib/formatters'

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
  onApply: () => void
}

const PRESETS: { id: TimePreset; label: string }[] = [
  { id: '1h', label: '1 Jam' },
  { id: '6h', label: '6 Jam' },
  { id: 'today', label: 'Hari Ini' },
  { id: '24h', label: '24 Jam' },
  { id: '7d', label: '7 Hari' },
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
  onApply,
}: PingAnalyticsControlsProps) {
  const [popoverOpen, setPopoverOpen] = useState(false)

  // Derive DateRange object from startDate & endDate
  const dateRange: DateRange | undefined = {
    from: startDate ? parseISO(startDate) : undefined,
    to: endDate ? parseISO(endDate) : undefined,
  }

  const handleDateRangeSelect = (range: DateRange | undefined) => {
    if (!range) return
    if (range.from) {
      onStartDateChange(formatDateInput(range.from))
    }
    if (range.to) {
      onEndDateChange(formatDateInput(range.to))
      onPresetChange('custom')
      setPopoverOpen(false)
    } else if (range.from) {
      onEndDateChange(formatDateInput(range.from))
      onPresetChange('custom')
    }
  }

  return (
    <div className='p-4 bg-muted/20 border-b flex flex-wrap items-center justify-between gap-3'>
      {/* Left section: Quick Presets & Date Range Popover */}
      <div className='flex flex-wrap items-center gap-2'>
        {/* Quick Presets */}
        <div className='flex items-center gap-1 bg-background border rounded-md p-0.5 shadow-2xs'>
          {PRESETS.map((p) => (
            <Button
              key={p.id}
              variant={selectedPreset === p.id ? 'secondary' : 'ghost'}
              size='sm'
              className='h-7 text-xs px-2.5 font-medium'
              onClick={() => onPresetChange(p.id)}
            >
              {p.label}
            </Button>
          ))}
        </div>

        {/* Shadcn Date Range Picker Popover */}
        <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
          <PopoverTrigger asChild>
            <Button
              id='date-picker-range'
              variant='outline'
              size='sm'
              className='h-8 text-xs font-normal justify-start gap-2 bg-background shadow-2xs'
            >
              <CalendarIcon className='h-3.5 w-3.5 text-muted-foreground' />
              {dateRange?.from ? (
                dateRange.to ? (
                  <span>
                    {format(dateRange.from, 'dd MMM yyyy')} -{' '}
                    {format(dateRange.to, 'dd MMM yyyy')}
                  </span>
                ) : (
                  format(dateRange.from, 'dd MMM yyyy')
                )
              ) : (
                <span className='text-muted-foreground'>Pilih Rentang Tanggal</span>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className='w-auto p-0' align='start'>
            <Calendar
              mode='range'
              defaultMonth={dateRange?.from}
              selected={dateRange}
              onSelect={handleDateRangeSelect}
              numberOfMonths={2}
            />
          </PopoverContent>
        </Popover>

        {/* Time Inputs */}
        <div className='flex items-center gap-1.5 bg-background border rounded-md px-2.5 py-1 text-xs shadow-2xs'>
          <Clock className='h-3.5 w-3.5 text-muted-foreground' />
          <input
            type='time'
            value={startTime}
            onChange={(e) => {
              onStartTimeChange(e.target.value)
              onPresetChange('custom')
            }}
            className='bg-transparent border-0 text-xs font-mono focus:outline-hidden w-14'
            title='Jam Mulai'
          />
          <span className='text-muted-foreground text-[10px]'>s/d</span>
          <input
            type='time'
            value={endTime}
            onChange={(e) => {
              onEndTimeChange(e.target.value)
              onPresetChange('custom')
            }}
            className='bg-transparent border-0 text-xs font-mono focus:outline-hidden w-14'
            title='Jam Selesai'
          />
        </div>
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
        <Button size='sm' className='h-8 text-xs' onClick={onApply}>
          Terapkan
        </Button>
      </div>
    </div>
  )
}
