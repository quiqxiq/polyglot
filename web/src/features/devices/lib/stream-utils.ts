/**
 * Check if a ConnectRPC / fetch stream error was intentionally aborted or cancelled.
 */
export function isStreamAbortedError(
  err: unknown,
  signal?: AbortSignal,
  active?: boolean
): boolean {
  if (active === false || signal?.aborted) return true
  if (!err) return false

  if (typeof err === 'object') {
    const errorObj = err as {
      name?: string
      code?: number
      message?: string
      rawMessage?: string
    }

    const name = errorObj.name || ''
    const msg = (errorObj.message || '').toLowerCase()
    const rawMsg = (errorObj.rawMessage || '').toLowerCase()

    return (
      name === 'AbortError' ||
      name === 'Canceled' ||
      errorObj.code === 1 || // ConnectRPC Canceled code
      msg.includes('aborted') ||
      msg.includes('input stream') ||
      msg.includes('canceled') ||
      msg.includes('cancelled') ||
      rawMsg.includes('input stream') ||
      rawMsg.includes('canceled') ||
      rawMsg.includes('cancelled')
    )
  }

  return false
}

