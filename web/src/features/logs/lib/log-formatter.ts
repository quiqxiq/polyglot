import { LogItem, LogSeverity } from '../types'

/**
 * Classifies log severity based on RouterOS topics and message content.
 * Error -> Red
 * Warning -> Yellow/Amber
 * Info -> Neutral/Blue
 * Debug -> Muted
 */
export function classifySeverity(topics: string, message: string): LogSeverity {
  const t = topics.toLowerCase()
  const m = message.toLowerCase()

  if (
    t.includes('error') ||
    t.includes('critical') ||
    m.includes('failure') ||
    m.includes('failed') ||
    m.includes('denied') ||
    m.includes('rejected') ||
    m.includes('unreachable') ||
    m.includes('kernel error')
  ) {
    return 'error'
  }

  if (
    t.includes('warning') ||
    t.includes('warn') ||
    m.includes('warn') ||
    m.includes('limit reached') ||
    m.includes('exceeded') ||
    m.includes('timeout') ||
    m.includes('disconnected')
  ) {
    return 'warning'
  }

  if (t.includes('debug')) {
    return 'debug'
  }

  return 'info'
}

/**
 * Exports given logs as a formatted TXT file download.
 */
export function exportLogsToTxt(logs: LogItem[], filename: string) {
  const content = logs
    .map((l) => `[${l.time}] [${l.severity.toUpperCase()}] [${l.topics || 'system'}] ${l.message}`)
    .join('\n')

  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

/**
 * Exports given logs as a JSON file download.
 */
export function exportLogsToJson(logs: LogItem[], filename: string) {
  const content = JSON.stringify(logs, null, 2)
  const blob = new Blob([content], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
