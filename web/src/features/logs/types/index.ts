export type LogSeverity = 'error' | 'warning' | 'info' | 'debug'

export type SeverityFilter = 'all' | 'error' | 'warning' | 'info'

export type LogTabType = 'all' | 'hotspot' | 'ppp'

export interface LogItem {
  id: string
  time: string
  topics: string
  message: string
  severity: LogSeverity
  timestamp: number
}
