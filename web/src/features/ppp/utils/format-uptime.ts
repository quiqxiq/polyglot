/**
 * Utility functions for parsing and formatting MikroTik RouterOS uptime strings
 * with local browser-side ticking.
 */

/**
 * Parses a RouterOS uptime string into total seconds.
 * Supports formats like:
 * - "22h35m27s", "35m27s", "45s"
 * - "2d15h30m10s", "1w2d03h04m05s"
 * - "1d04:12:30", "04:12:30"
 */
export function parseUptimeToSeconds(uptime?: string): number {
  if (!uptime || uptime === 'N/A' || uptime === '-' || uptime === '') return 0

  let weeks = 0
  let days = 0
  let hours = 0
  let minutes = 0
  let seconds = 0

  // Check for week notation (e.g. 1w)
  const wMatch = uptime.match(/(\d+)w/)
  if (wMatch) weeks = parseInt(wMatch[1], 10)

  // Check for colon-separated format (e.g. "1d04:12:30" or "04:12:30")
  const colonMatch = uptime.match(/(?:(\d+)d)?(?:(\d+)h)?(\d{1,2}):(\d{2}):(\d{2})/)
  if (colonMatch) {
    if (colonMatch[1]) days = parseInt(colonMatch[1], 10)
    if (colonMatch[2]) hours = parseInt(colonMatch[2], 10)
    hours = parseInt(colonMatch[3], 10)
    minutes = parseInt(colonMatch[4], 10)
    seconds = parseInt(colonMatch[5], 10)
  } else {
    // Check standard RouterOS letter notation (e.g. 2d15h30m10s)
    const dMatch = uptime.match(/(\d+)d/)
    const hMatch = uptime.match(/(\d+)h/)
    const mMatch = uptime.match(/(\d+)m/)
    const sMatch = uptime.match(/(\d+)s/)
    if (dMatch) days = parseInt(dMatch[1], 10)
    if (hMatch) hours = parseInt(hMatch[1], 10)
    if (mMatch) minutes = parseInt(mMatch[1], 10)
    if (sMatch) seconds = parseInt(sMatch[1], 10)
  }

  return weeks * 604800 + days * 86400 + hours * 3600 + minutes * 60 + seconds
}

/**
 * Formats total seconds into standard uptime display string:
 * - When days > 0: "1 hari 04:12:30" (or "2 hari 15:30:10")
 * - When days == 0: "22:35:27"
 */
export function formatSecondsToUptime(totalSeconds: number): string {
  if (totalSeconds <= 0) return '00:00:00'

  const days = Math.floor(totalSeconds / 86400)
  const rem = totalSeconds % 86400
  const hours = Math.floor(rem / 3600)
  const minutes = Math.floor((rem % 3600) / 60)
  const seconds = rem % 60

  const hStr = hours.toString().padStart(2, '0')
  const mStr = minutes.toString().padStart(2, '0')
  const sStr = seconds.toString().padStart(2, '0')
  const timeStr = `${hStr}:${mStr}:${sStr}`

  if (days > 0) {
    return `${days} hari ${timeStr}`
  }
  return timeStr
}

/**
 * Formats a raw RouterOS uptime string with optional added elapsed seconds.
 */
export function formatUptime(uptime?: string, elapsedSeconds = 0): string {
  if (!uptime || uptime === '-' || uptime === 'N/A' || uptime === '') return '-'
  const baseSeconds = parseUptimeToSeconds(uptime)
  if (baseSeconds === 0 && !uptime.includes('0s')) return uptime
  return formatSecondsToUptime(baseSeconds + elapsedSeconds)
}
