import { RefreshCw, SearchIcon } from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ONU_STATE_LABEL_KEYS, ONU_STATE_OPTIONS } from '../../lib/constants'

export const ALL_FILTER = 'ALL'

interface OltOnuFiltersProps {
  ponCount: number
  ponIndex: string
  state: string
  search: string
  isFetching: boolean
  onPonChange: (value: string) => void
  onStateChange: (value: string) => void
  onSearchChange: (value: string) => void
  onRefresh: () => void
}

export function OltOnuFilters({
  ponCount,
  ponIndex,
  state,
  search,
  isFetching,
  onPonChange,
  onStateChange,
  onSearchChange,
  onRefresh,
}: OltOnuFiltersProps) {
  const { t } = useI18n()
  const ponOptions = Array.from({ length: Math.max(ponCount, 1) }, (_, index) =>
    String(index + 1)
  )

  return (
    <div className='flex flex-wrap items-center gap-2'>
      <div className='relative w-full sm:max-w-xs'>
        <SearchIcon className='absolute start-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
        <Input
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder={t('olt.search_onu_placeholder')}
          className='ps-8'
        />
      </div>

      <Select value={ponIndex} onValueChange={onPonChange}>
        <SelectTrigger className='w-36' aria-label={t('olt.filter_pon')}>
          <SelectValue placeholder={t('olt.filter_pon')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_FILTER}>{t('olt.all_pon')}</SelectItem>
          {ponOptions.map((value) => (
            <SelectItem key={value} value={value}>
              PON {value}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select value={state} onValueChange={onStateChange}>
        <SelectTrigger className='w-40' aria-label={t('olt.filter_state')}>
          <SelectValue placeholder={t('olt.filter_state')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_FILTER}>{t('olt.all_state')}</SelectItem>
          {ONU_STATE_OPTIONS.map((value) => (
            <SelectItem key={value} value={value}>
              {t(ONU_STATE_LABEL_KEYS[value] ?? 'olt.state_unknown')}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Button
        variant='outline'
        size='sm'
        className='h-9'
        onClick={onRefresh}
        disabled={isFetching}
      >
        <RefreshCw
          className={`mr-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`}
        />
        {t('common.refresh')}
      </Button>
    </div>
  )
}
