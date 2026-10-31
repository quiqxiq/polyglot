export const OLT_TRANSPORT_OPTIONS = [
  { value: 'telnet', label: 'Telnet' },
  { value: 'ssh', label: 'SSH' },
] as const

export const OLT_VARIANT_OPTIONS = ['AUTO', 'HA7302', 'HA7304'] as const

export const ONU_STATE_OPTIONS = [
  'ONLINE',
  'OFFLINE',
  'UNKNOWN',
  'REMOVED',
] as const

export const ONU_STATE_LABEL_KEYS: Record<string, string> = {
  ONLINE: 'olt.state_online',
  OFFLINE: 'olt.state_offline',
  UNKNOWN: 'olt.state_unknown',
  REMOVED: 'olt.state_removed',
}

const EMERALD =
  'border-emerald-500/30 bg-emerald-500/15 text-emerald-700 dark:text-emerald-400'
const AMBER =
  'border-amber-500/30 bg-amber-500/15 text-amber-700 dark:text-amber-400'
const ROSE =
  'border-rose-500/30 bg-rose-500/15 text-rose-700 dark:text-rose-400'
const MUTED = 'border-muted-foreground/30 bg-muted text-muted-foreground'

export function enabledBadgeClass(enabled: boolean): string {
  return enabled ? EMERALD : MUTED
}

export function onuStateBadgeClass(state: string): string {
  switch (state) {
    case 'ONLINE':
      return EMERALD
    case 'OFFLINE':
      return ROSE
    case 'UNKNOWN':
      return AMBER
    default:
      return MUTED
  }
}

export function rxClassBadgeClass(rxClass: string): string {
  switch (rxClass) {
    case 'NORMAL':
      return EMERALD
    case 'WARNING':
      return AMBER
    case 'CRITICAL':
      return ROSE
    default:
      return MUTED
  }
}

export function rxClassLabelKey(rxClass: string): string {
  switch (rxClass) {
    case 'NORMAL':
      return 'olt.rx_class_normal'
    case 'WARNING':
      return 'olt.rx_class_warning'
    case 'CRITICAL':
      return 'olt.rx_class_critical'
    default:
      return 'olt.rx_class_unavailable'
  }
}

export function rxClassTextClass(rxClass: string): string {
  switch (rxClass) {
    case 'WARNING':
      return 'text-amber-600 dark:text-amber-400'
    case 'CRITICAL':
      return 'text-rose-600 dark:text-rose-400'
    case 'UNAVAILABLE':
      return 'text-muted-foreground'
    default:
      return 'text-foreground'
  }
}

export function onuPositionLabel(ponIndex: number, onuIndex: number): string {
  return `PON ${ponIndex} / ONU ${onuIndex}`
}
