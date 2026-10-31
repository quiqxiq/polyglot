import type { OltHealth } from '@/gen/v1/olt_pb'
import { cn } from '@/lib/utils'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import {
  enabledBadgeClass,
  onuStateBadgeClass,
  ONU_STATE_LABEL_KEYS,
  rxClassBadgeClass,
  rxClassLabelKey,
} from '../lib/constants'

export function EnabledBadge({ enabled }: { enabled: boolean }) {
  const { t } = useI18n()
  return (
    <Badge variant='outline' className={cn(enabledBadgeClass(enabled))}>
      {enabled ? t('olt.admin_enabled') : t('olt.admin_disabled')}
    </Badge>
  )
}

export function OnuStateBadge({ state }: { state: string }) {
  const { t } = useI18n()
  const labelKey = ONU_STATE_LABEL_KEYS[state] ?? 'olt.state_unknown'
  return (
    <Badge variant='outline' className={cn(onuStateBadgeClass(state))}>
      {t(labelKey)}
    </Badge>
  )
}

export function RxClassBadge({ rxClass }: { rxClass: string }) {
  const { t } = useI18n()
  return (
    <Badge variant='outline' className={cn(rxClassBadgeClass(rxClass))}>
      {t(rxClassLabelKey(rxClass))}
    </Badge>
  )
}

export function HealthBadge({ health }: { health?: OltHealth }) {
  const { t } = useI18n()
  const reachable = Boolean(health?.reachable)
  return (
    <Badge
      variant='outline'
      className={cn(
        reachable
          ? 'border-emerald-500/30 bg-emerald-500/15 text-emerald-700 dark:text-emerald-400'
          : 'border-rose-500/30 bg-rose-500/15 text-rose-700 dark:text-rose-400'
      )}
    >
      {reachable ? t('olt.reachable') : t('olt.unreachable')}
    </Badge>
  )
}
