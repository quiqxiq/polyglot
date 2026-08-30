/* eslint-disable max-lines */
'use client'

import { type FC, useState, useEffect, useMemo, JSX } from 'react'
import { Button } from './button'
import { Popover, PopoverContent, PopoverTrigger } from './popover'
import { Calendar } from './calendar'
import { Label } from './label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './select'
import { Switch } from './switch'
import { Calendar as CalendarIcon, Check, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { DateInput } from './date-input'
import { TimeInput } from './time-input'

export interface DateRange {
  from: Date
  to: Date | undefined
}

export interface DateRangePreset {
  name: string
  label: string
  getRange?: () => DateRange
}

export interface DateRangePickerProps {
  /** Click handler for applying the updates from DateRangePicker. */
  onUpdate?: (values: { range: DateRange; rangeCompare?: DateRange; preset?: string }) => void
  /** Initial value for start date */
  initialDateFrom?: Date | string
  /** Initial value for end date */
  initialDateTo?: Date | string
  /** Initial value for start date for compare */
  initialCompareFrom?: Date | string
  /** Initial value for end date for compare */
  initialCompareTo?: Date | string
  /** Alignment of popover */
  align?: 'start' | 'center' | 'end'
  /** Option for locale */
  locale?: string
  /** Option for showing compare feature */
  showCompare?: boolean
  /** Option for showing time picker */
  showTimePicker?: boolean
  /** List of presets to display. If empty, presets sidebar/select is omitted. */
  presets?: DateRangePreset[]
  /** Trigger button className */
  className?: string
  /** Trigger button size */
  size?: 'default' | 'sm' | 'lg' | 'icon'
  /** Trigger button variant */
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  /** Disabled state */
  disabled?: boolean
  /** Selected preset */
  selectedPreset?: string
  /** Placeholder text */
  placeholder?: string
}

const formatDate = (date: Date, locale: string = 'id-ID', includeTime: boolean = false): string => {
  if (isNaN(date.getTime())) return ''
  const options: Intl.DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    ...(includeTime
      ? {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
        }
      : {}),
  }
  return date.toLocaleDateString(locale, options)
}

const getDateAdjustedForTimezone = (dateInput: Date | string): Date => {
  if (typeof dateInput === 'string') {
    if (dateInput.includes('T') || dateInput.includes(' ')) {
      const parsed = new Date(dateInput)
      if (!isNaN(parsed.getTime())) return parsed
    }
    const parts = dateInput.split('-').map((part) => parseInt(part, 10))
    if (parts.length >= 3) {
      return new Date(parts[0], parts[1] - 1, parts[2], 0, 0, 0)
    }
    const fallback = new Date(dateInput)
    return isNaN(fallback.getTime()) ? new Date() : fallback
  }
  return dateInput
}

const DateRangePicker: FC<DateRangePickerProps> = ({
  initialDateFrom = new Date(),
  initialDateTo,
  initialCompareFrom,
  initialCompareTo,
  onUpdate,
  align = 'end',
  locale = 'id-ID',
  showCompare = false,
  showTimePicker = false,
  presets = [],
  className,
  size = 'default',
  variant = 'outline',
  disabled = false,
  selectedPreset: controlledPreset,
  placeholder = 'Pilih rentang tanggal',
}): JSX.Element => {
  const [isOpen, setIsOpen] = useState(false)

  const [range, setRange] = useState<DateRange>(() => {
    const from = getDateAdjustedForTimezone(initialDateFrom)
    const to = initialDateTo
      ? getDateAdjustedForTimezone(initialDateTo)
      : getDateAdjustedForTimezone(initialDateFrom)
    return { from, to }
  })

  const [rangeCompare, setRangeCompare] = useState<DateRange | undefined>(() => {
    if (!initialCompareFrom) return undefined
    const from = getDateAdjustedForTimezone(initialCompareFrom)
    const to = initialCompareTo
      ? getDateAdjustedForTimezone(initialCompareTo)
      : getDateAdjustedForTimezone(initialCompareFrom)
    return { from, to }
  })

  // Synchronize internal state when initial props change
  useEffect(() => {
    if (initialDateFrom) {
      setRange((prev) => {
        const nextFrom = getDateAdjustedForTimezone(initialDateFrom)
        const nextTo = initialDateTo ? getDateAdjustedForTimezone(initialDateTo) : nextFrom
        if (
          prev.from.getTime() === nextFrom.getTime() &&
          prev.to?.getTime() === nextTo.getTime()
        ) {
          return prev
        }
        return { from: nextFrom, to: nextTo }
      })
    }
  }, [initialDateFrom, initialDateTo])

  const [selectedPreset, setSelectedPreset] = useState<string | undefined>(controlledPreset)

  useEffect(() => {
    if (controlledPreset !== undefined) {
      setSelectedPreset(controlledPreset)
    }
  }, [controlledPreset])

  const [isSmallScreen, setIsSmallScreen] = useState(
    typeof window !== 'undefined' ? window.innerWidth < 960 : false
  )

  useEffect(() => {
    const handleResize = (): void => {
      setIsSmallScreen(window.innerWidth < 960)
    }
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  const hasPresets = Boolean(presets && presets.length > 0)

  const getPresetRange = (presetName: string): DateRange => {
    const custom = presets.find((p) => p.name === presetName)
    if (custom?.getRange) {
      return custom.getRange()
    }

    const now = new Date()
    const from = new Date()
    const to = new Date()

    switch (presetName) {
      case '1h':
        from.setTime(now.getTime() - 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '6h':
        from.setTime(now.getTime() - 6 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '12h':
        from.setTime(now.getTime() - 12 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '24h':
        from.setTime(now.getTime() - 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case 'today':
        from.setHours(0, 0, 0, 0)
        to.setTime(now.getTime())
        break
      case 'yesterday':
        from.setDate(from.getDate() - 1)
        from.setHours(0, 0, 0, 0)
        to.setDate(to.getDate() - 1)
        to.setHours(23, 59, 59, 999)
        break
      case '3d':
        from.setTime(now.getTime() - 3 * 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '7d':
      case 'last7':
        from.setTime(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '14d':
      case 'last14':
        from.setTime(now.getTime() - 14 * 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '15d':
        from.setTime(now.getTime() - 15 * 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case '30d':
      case 'last30':
        from.setTime(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        to.setTime(now.getTime())
        break
      case 'thisWeek': {
        const first = from.getDate() - from.getDay()
        from.setDate(first)
        from.setHours(0, 0, 0, 0)
        to.setTime(now.getTime())
        break
      }
      case 'lastWeek': {
        from.setDate(from.getDate() - 7 - from.getDay())
        from.setHours(0, 0, 0, 0)
        to.setDate(to.getDate() - to.getDay() - 1)
        to.setHours(23, 59, 59, 999)
        break
      }
      case 'thisMonth':
        from.setDate(1)
        from.setHours(0, 0, 0, 0)
        to.setTime(now.getTime())
        break
      case 'lastMonth':
        from.setMonth(from.getMonth() - 1)
        from.setDate(1)
        from.setHours(0, 0, 0, 0)
        to.setDate(0)
        to.setHours(23, 59, 59, 999)
        break
      default:
        from.setHours(0, 0, 0, 0)
        to.setHours(23, 59, 59, 999)
    }

    return { from, to }
  }

  const setPreset = (presetName: string): void => {
    const newRange = getPresetRange(presetName)
    setRange(newRange)
    setSelectedPreset(presetName)

    let compare: DateRange | undefined
    if (rangeCompare) {
      compare = {
        from: new Date(
          newRange.from.getFullYear() - 1,
          newRange.from.getMonth(),
          newRange.from.getDate(),
          newRange.from.getHours(),
          newRange.from.getMinutes()
        ),
        to: newRange.to
          ? new Date(
              newRange.to.getFullYear() - 1,
              newRange.to.getMonth(),
              newRange.to.getDate(),
              newRange.to.getHours(),
              newRange.to.getMinutes()
            )
          : undefined,
      }
      setRangeCompare(compare)
    }
    onUpdate?.({ range: newRange, rangeCompare: compare, preset: presetName })
    setIsOpen(false)
  }

  const updateFromTime = (time: { hours: number; minutes: number; seconds: number }) => {
    setSelectedPreset('custom')
    const updatedFrom = new Date(range.from)
    updatedFrom.setHours(time.hours, time.minutes, time.seconds, 0)
    const newRange = { ...range, from: updatedFrom }
    setRange(newRange)
    onUpdate?.({ range: newRange, rangeCompare, preset: 'custom' })
  }

  const updateToTime = (time: { hours: number; minutes: number; seconds: number }) => {
    setSelectedPreset('custom')
    const baseTo = range.to ? new Date(range.to) : new Date(range.from)
    baseTo.setHours(time.hours, time.minutes, time.seconds, 0)
    const newRange = { ...range, to: baseTo }
    setRange(newRange)
    onUpdate?.({ range: newRange, rangeCompare, preset: 'custom' })
  }

  const formattedDisplay = useMemo(() => {
    if (!range.from) return placeholder
    const fromStr = formatDate(range.from, locale, showTimePicker)
    if (!range.to || range.from.getTime() === range.to.getTime()) {
      return fromStr
    }
    const toStr = formatDate(range.to, locale, showTimePicker)
    return `${fromStr} – ${toStr}`
  }, [range, locale, showTimePicker, placeholder])

  const PresetButton = ({
    preset,
    label,
    isSelected,
  }: {
    preset: string
    label: string
    isSelected: boolean
  }): JSX.Element => (
    <Button
      type='button'
      size='sm'
      variant={isSelected ? 'secondary' : 'ghost'}
      className={cn(
        'w-full justify-start text-xs font-normal h-8 px-2.5',
        isSelected && 'font-medium bg-primary/10 text-primary hover:bg-primary/15'
      )}
      onClick={() => setPreset(preset)}
    >
      <span className={cn('mr-1.5 flex h-3.5 w-3.5 items-center justify-center opacity-0', isSelected && 'opacity-100')}>
        <Check className='h-3.5 w-3.5' />
      </span>
      {label}
    </Button>
  )

  return (
    <Popover modal={true} open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          size={size}
          variant={variant}
          disabled={disabled}
          className={cn(
            'h-9 justify-start text-xs font-normal gap-2 shadow-2xs',
            className
          )}
        >
          <CalendarIcon className='h-3.5 w-3.5 text-muted-foreground shrink-0' />
          <span className='truncate'>{formattedDisplay}</span>
          {rangeCompare != null && (
            <span className='opacity-60 text-[10px] truncate'>
              (vs {formatDate(rangeCompare.from, locale, showTimePicker)})
            </span>
          )}
          <div className='ml-auto pl-1 opacity-60'>
            {isOpen ? <ChevronUp className='h-3.5 w-3.5' /> : <ChevronDown className='h-3.5 w-3.5' />}
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent align={align} className='w-auto p-0 shadow-lg border' sideOffset={4}>
        <div className='flex flex-col md:flex-row divide-y md:divide-y-0 md:divide-x'>
          {/* Main Picker Body */}
          <div className='flex flex-col p-3 gap-3'>
            {/* Header: Date & Time inputs */}
            <div className='flex flex-col gap-2'>
              {showCompare && (
                <div className='flex items-center space-x-2 pb-1'>
                  <Switch
                    checked={Boolean(rangeCompare)}
                    onCheckedChange={(checked: boolean) => {
                      if (checked) {
                        const newCompare = {
                          from: new Date(
                            range.from.getFullYear() - 1,
                            range.from.getMonth(),
                            range.from.getDate()
                          ),
                          to: range.to
                            ? new Date(
                                range.to.getFullYear() - 1,
                                range.to.getMonth(),
                                range.to.getDate()
                              )
                            : new Date(
                                range.from.getFullYear() - 1,
                                range.from.getMonth(),
                                range.from.getDate()
                              ),
                        }
                        setRangeCompare(newCompare)
                        onUpdate?.({ range, rangeCompare: newCompare, preset: selectedPreset })
                      } else {
                        setRangeCompare(undefined)
                        onUpdate?.({ range, rangeCompare: undefined, preset: selectedPreset })
                      }
                    }}
                    id='compare-mode'
                  />
                  <Label htmlFor='compare-mode' className='text-xs'>
                    Bandingkan Periode
                  </Label>
                </div>
              )}

              {/* Range Inputs (Start & End with Time) */}
              <div className='flex flex-wrap items-center gap-2 bg-muted/40 p-2 rounded-lg border'>
                <div className='flex items-center gap-1.5'>
                  <span className='text-[11px] font-semibold text-muted-foreground w-8'>Dari:</span>
                  <DateInput
                    value={range.from}
                    onChange={(date) => {
                      setSelectedPreset('custom')
                      const hours = range.from.getHours()
                      const minutes = range.from.getMinutes()
                      const seconds = range.from.getSeconds()
                      date.setHours(hours, minutes, seconds, 0)
                      const toDate = range.to == null || date > range.to ? date : range.to
                      const newRange = { from: date, to: toDate }
                      setRange(newRange)
                      onUpdate?.({ range: newRange, rangeCompare, preset: 'custom' })
                    }}
                  />
                  {showTimePicker && (
                    <TimeInput
                      value={range.from}
                      onTimeChange={updateFromTime}
                    />
                  )}
                </div>

                <span className='text-muted-foreground hidden sm:inline'>–</span>

                <div className='flex items-center gap-1.5'>
                  <span className='text-[11px] font-semibold text-muted-foreground w-8 sm:hidden'>S/d:</span>
                  <DateInput
                    value={range.to ?? range.from}
                    onChange={(date) => {
                      setSelectedPreset('custom')
                      const hours = range.to ? range.to.getHours() : 23
                      const minutes = range.to ? range.to.getMinutes() : 59
                      const seconds = range.to ? range.to.getSeconds() : 59
                      date.setHours(hours, minutes, seconds, 0)
                      const fromDate = date < range.from ? date : range.from
                      const newRange = { from: fromDate, to: date }
                      setRange(newRange)
                      onUpdate?.({ range: newRange, rangeCompare, preset: 'custom' })
                    }}
                  />
                  {showTimePicker && (
                    <TimeInput
                      value={range.to ?? range.from}
                      onTimeChange={updateToTime}
                    />
                  )}
                </div>
              </div>
            </div>

            {/* Responsive preset dropdown for small mobile screen */}
            {isSmallScreen && hasPresets && (
              <Select
                value={selectedPreset}
                onValueChange={(value) => {
                  setPreset(value)
                }}
              >
                <SelectTrigger className='w-full text-xs h-8'>
                  <SelectValue placeholder='Pilih Preset Cepat...' />
                </SelectTrigger>
                <SelectContent>
                  {presets.map((preset) => (
                    <SelectItem key={preset.name} value={preset.name} className='text-xs'>
                      {preset.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            {/* Calendar Widget */}
            <div className='rounded-md border bg-background p-1 flex justify-center'>
              <Calendar
                mode='range'
                onSelect={(value: { from?: Date; to?: Date } | undefined) => {
                  setSelectedPreset('custom')
                  if (value?.from != null) {
                    const nextFrom = new Date(value.from)
                    nextFrom.setHours(
                      range.from.getHours(),
                      range.from.getMinutes(),
                      range.from.getSeconds()
                    )

                    let nextTo: Date | undefined
                    if (value.to) {
                      nextTo = new Date(value.to)
                      const toHours = range.to ? range.to.getHours() : 23
                      const toMinutes = range.to ? range.to.getMinutes() : 59
                      const toSeconds = range.to ? range.to.getSeconds() : 59
                      nextTo.setHours(toHours, toMinutes, toSeconds)
                    }
                    const nextRange = { from: nextFrom, to: nextTo }
                    setRange(nextRange)
                    if (nextTo) {
                      onUpdate?.({ range: nextRange, rangeCompare, preset: 'custom' })
                    }
                  }
                }}
                selected={range}
                numberOfMonths={isSmallScreen ? 1 : 2}
                defaultMonth={range.from}
              />
            </div>
          </div>

          {/* Desktop Presets Sidebar (only rendered if presets are provided) */}
          {!isSmallScreen && hasPresets && (
            <div className='flex w-44 flex-col gap-1 p-3 bg-muted/20 overflow-y-auto max-h-[380px]'>
              <div className='px-2 py-1 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider'>
                Preset Waktu
              </div>
              <div className='flex flex-col gap-0.5'>
                {presets.map((preset) => (
                  <PresetButton
                    key={preset.name}
                    preset={preset.name}
                    label={preset.label}
                    isSelected={selectedPreset === preset.name}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}

DateRangePicker.displayName = 'DateRangePicker'

export { DateRangePicker }