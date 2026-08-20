export const PPP_SERVICE_OPTIONS = [
  { label: 'Any', value: 'any' },
  { label: 'PPPoE', value: 'pppoe' },
  { label: 'L2TP', value: 'l2tp' },
  { label: 'PPTP', value: 'pptp' },
  { label: 'SSTP', value: 'sstp' },
  { label: 'OpenVPN', value: 'ovpn' },
] as const

export const PPP_ONLY_ONE_OPTIONS = [
  { label: 'Default', value: 'default' },
  { label: 'Yes (Enforce 1 session)', value: 'yes' },
  { label: 'No (Multiple sessions)', value: 'no' },
] as const

export const PPP_RATE_LIMIT_PRESETS = [
  { label: '2 Mbps (2M/2M)', value: '2M/2M' },
  { label: '5 Mbps (5M/5M)', value: '5M/5M' },
  { label: '10 Mbps (10M/10M)', value: '10M/10M' },
  { label: '20 Mbps (20M/20M)', value: '20M/20M' },
  { label: '50 Mbps (50M/50M)', value: '50M/50M' },
  { label: '100 Mbps (100M/100M)', value: '100M/100M' },
  { label: 'Isolir / Suspended (0/0)', value: '0/0' },
] as const
