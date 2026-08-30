import * as React from 'react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

export interface TimeInputProps
  extends Omit<React.ComponentProps<typeof Input>, 'value' | 'onChange'> {
  value?: Date | string
  onChange?: (val: string) => void
  onTimeChange?: (time: { hours: number; minutes: number; seconds: number }) => void
}

function parseTimeToValue(val?: Date | string): string {
  if (!val) return '00:00'
  if (val instanceof Date) {
    if (isNaN(val.getTime())) return '00:00'
    const hh = String(val.getHours()).padStart(2, '0')
    const mm = String(val.getMinutes()).padStart(2, '0')
    return `${hh}:${mm}`
  }
  if (typeof val === 'string') {
    if (val.includes('T')) {
      const d = new Date(val)
      if (!isNaN(d.getTime())) {
        const hh = String(d.getHours()).padStart(2, '0')
        const mm = String(d.getMinutes()).padStart(2, '0')
        return `${hh}:${mm}`
      }
    }
    const parts = val.split(':')
    if (parts.length >= 2) {
      const hh = parts[0].padStart(2, '0')
      const mm = parts[1].padStart(2, '0')
      return `${hh}:${mm}`
    }
    return val
  }
  return '00:00'
}

const TimeInput = React.forwardRef<HTMLInputElement, TimeInputProps>(
  ({ className, value, onChange, onTimeChange, disabled, ...props }, ref) => {
    const timeStr = React.useMemo(() => parseTimeToValue(value), [value])

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value
      onChange?.(val)
      if (val && val.includes(':')) {
        const [h, m] = val.split(':').map((n) => parseInt(n, 10))
        onTimeChange?.({
          hours: isNaN(h) ? 0 : h,
          minutes: isNaN(m) ? 0 : m,
          seconds: 0,
        })
      }
    }

    return (
      <Input
        ref={ref}
        type='time'
        value={timeStr}
        onChange={handleChange}
        disabled={disabled}
        className={cn('h-8 w-24 px-2 text-xs font-mono', className)}
        {...props}
      />
    )
  }
)

TimeInput.displayName = 'TimeInput'

export { TimeInput }
