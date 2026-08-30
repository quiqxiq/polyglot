import { useState, useMemo } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Copy, Check, Table2, Terminal } from 'lucide-react'
import { toast } from 'sonner'
import type { PingMetricPointData } from '@/gen/v1/device_pb'

interface PingAnalyticsTableProps {
  points: PingMetricPointData[]
}

interface RawPingBlock {
  type: 'header' | 'row' | 'summary'
  content: string
  isTimeout?: boolean
  summaryData?: {
    sent: number
    received: number
    lossPct: number
    minRtt: number
    avgRtt: number
    maxRtt: number
  }
}

export function PingAnalyticsTable({ points }: PingAnalyticsTableProps) {
  const [viewMode, setViewMode] = useState<'table' | 'raw'>('table')
  const [copied, setCopied] = useState(false)

  const displayPoints = useMemo(() => {
    return points.slice(-200)
  }, [points])

  // Generate RouterOS CLI formatted blocks
  const { rawBlocks, rawText } = useMemo(() => {
    if (!points || points.length === 0) {
      return {
        rawBlocks: [{ type: 'row' as const, content: '  Tidak ada log paket pada rentang ini.' }],
        rawText: 'Tidak ada log paket pada rentang ini.',
      }
    }

    const blocks: RawPingBlock[] = []
    const rawLines: string[] = []
    const headerStr = '  SEQ HOST                                     SIZE TTL TIME  STATUS'

    let totalSent = 0
    let totalReceived = 0
    let minRtt = Infinity
    let maxRtt = 0
    let rttSum = 0
    let rttCount = 0
    let currentBatch = 0

    blocks.push({ type: 'header', content: headerStr })
    rawLines.push(headerStr)

    for (let i = 0; i < points.length; i++) {
      const p = points[i]
      const sentDelta = p.sent > 0 ? p.sent : 1
      totalSent += sentDelta

      const isTimeout = p.status === 'timeout' || p.rttMs <= 0 || p.packetLoss >= 100
      if (!isTimeout) {
        const recvDelta = p.received > 0 ? p.received : 1
        totalReceived += recvDelta
        if (p.rttMs < minRtt) minRtt = p.rttMs
        if (p.rttMs > maxRtt) maxRtt = p.rttMs
        rttSum += p.rttMs
        rttCount++
      }

      const seqNum = p.seq > 0 ? p.seq : i
      const seqStr = String(seqNum).padStart(5, ' ')
      const hostStr = (p.target || '8.8.8.8').padEnd(41, ' ')
      const sizeStr = String(p.size || 56).padStart(4, ' ')
      const ttlStr = String(p.ttl || (isTimeout ? '-' : 116)).padStart(4, ' ')
      const timeStr = isTimeout ? 'timeout' : `${Math.round(p.rttMs)}ms`
      const timePadded = timeStr.padStart(5, ' ')
      const statusStr = isTimeout
        ? 'timeout'
        : p.status && p.status !== 'connected' && p.status !== 'ok'
        ? p.status
        : ''

      const rowLine = `${seqStr} ${hostStr} ${sizeStr} ${ttlStr} ${timePadded} ${statusStr}`.trimEnd()
      blocks.push({ type: 'row', content: rowLine, isTimeout })
      rawLines.push(rowLine)
      currentBatch++

      // Every 20 packets or at the very last packet
      if (currentBatch === 10 || i === points.length - 1) {
        const lossPct =
          totalSent > 0
            ? Math.max(0, Math.min(100, Math.round(((totalSent - totalReceived) / totalSent) * 100)))
            : 0
        const curMin = minRtt === Infinity ? 0 : Math.round(minRtt)
        const curAvg = rttCount > 0 ? Math.round(rttSum / rttCount) : 0
        const curMax = Math.round(maxRtt)

        const summaryLine = `    sent=${totalSent} received=${totalReceived} packet-loss=${lossPct}% min-rtt=${curMin}ms avg-rtt=${curAvg}ms max-rtt=${curMax}ms`
        blocks.push({
          type: 'summary',
          content: summaryLine,
          summaryData: {
            sent: totalSent,
            received: totalReceived,
            lossPct,
            minRtt: curMin,
            avgRtt: curAvg,
            maxRtt: curMax,
          },
        })
        rawLines.push(summaryLine)

        if (i < points.length - 1) {
          blocks.push({ type: 'header', content: headerStr })
          rawLines.push(headerStr)
          currentBatch = 0
        }
      }
    }

    return { rawBlocks: blocks, rawText: rawLines.join('\n') }
  }, [points])

  const handleCopyRaw = async () => {
    try {
      await navigator.clipboard.writeText(rawText)
      setCopied(true)
      toast.success('Log output CLI ping berhasil disalin')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Gagal menyalin log')
    }
  }

  return (
    <div className='border rounded-lg overflow-hidden bg-card shadow-2xs space-y-0'>
      {/* Control Header: View Mode Switch & Copy Button */}
      <div className='flex items-center justify-between px-3 py-2 border-b bg-muted/30'>
        <div className='flex items-center gap-1 bg-background border rounded-md p-0.5 shadow-2xs'>
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size='sm'
            className='h-6 text-[11px] px-2.5 gap-1.5 font-medium'
            onClick={() => setViewMode('table')}
          >
            <Table2 className='h-3.5 w-3.5' />
            Tabel
          </Button>
          <Button
            variant={viewMode === 'raw' ? 'secondary' : 'ghost'}
            size='sm'
            className='h-6 text-[11px] px-2.5 gap-1.5 font-medium'
            onClick={() => setViewMode('raw')}
          >
            <Terminal className='h-3.5 w-3.5' />
            Raw CLI Output
          </Button>
        </div>

        {viewMode === 'raw' && (
          <Button
            variant='outline'
            size='sm'
            className='h-6 text-[11px] px-2 gap-1 bg-background shadow-2xs'
            onClick={handleCopyRaw}
          >
            {copied ? (
              <>
                <Check className='h-3 w-3 text-emerald-500' />
                Tersalin
              </>
            ) : (
              <>
                <Copy className='h-3 w-3 text-muted-foreground' />
                Salin CLI
              </>
            )}
          </Button>
        )}
      </div>

      {/* Content: Table View vs Raw CLI View */}
      {viewMode === 'table' ? (
        <div className='max-h-72 overflow-y-auto'>
          <table className='w-full text-xs text-left'>
            <thead className='bg-muted/60 text-muted-foreground sticky top-0 border-b backdrop-blur-xs'>
              <tr>
                <th className='p-2.5 font-medium'>Waktu</th>
                <th className='p-2.5 font-medium'>SEQ</th>
                <th className='p-2.5 font-medium'>Host Target</th>
                <th className='p-2.5 font-medium'>Size</th>
                <th className='p-2.5 font-medium'>TTL</th>
                <th className='p-2.5 font-medium'>Latency</th>
                <th className='p-2.5 font-medium'>Status</th>
              </tr>
            </thead>
            <tbody className='divide-y font-mono text-[11px]'>
              {points.length === 0 ? (
                <tr>
                  <td
                    colSpan={7}
                    className='p-6 text-center text-muted-foreground font-sans text-xs'
                  >
                    Tidak ada log paket pada rentang ini.
                  </td>
                </tr>
              ) : (
                displayPoints
                  .slice()
                  .reverse()
                  .map((p, i) => {
                    const isSuccess =
                      p.status === 'connected' ||
                      p.status === 'ok' ||
                      (p.packetLoss === 0 && p.rttMs > 0)

                    return (
                      <tr
                        key={`${p.timestamp}_${p.seq}_${i}`}
                        className='hover:bg-muted/30 transition-colors'
                      >
                        <td className='p-2.5 text-muted-foreground'>
                          {p.timestamp
                            ? new Date(p.timestamp).toLocaleTimeString()
                            : '-'}
                        </td>
                        <td className='p-2.5 font-semibold text-foreground'>
                          {p.seq > 0 ? p.seq : displayPoints.length - 1 - i}
                        </td>
                        <td className='p-2.5 text-primary'>{p.target}</td>
                        <td className='p-2.5'>{p.size || 56}B</td>
                        <td className='p-2.5'>{p.ttl || '-'}</td>
                        <td className='p-2.5 font-semibold text-foreground'>
                          {p.rttMs > 0 ? `${p.rttMs.toFixed(1)} ms` : 'Timeout'}
                        </td>
                        <td className='p-2.5'>
                          <Badge
                            variant='outline'
                            className={`text-[10px] px-1.5 py-0 capitalize ${
                              isSuccess
                                ? 'bg-emerald-500/10 text-emerald-600 border-emerald-300 dark:border-emerald-800'
                                : 'bg-destructive/10 text-destructive border-destructive/30'
                            }`}
                          >
                            {p.status || (isSuccess ? 'Success' : 'Timeout')}
                          </Badge>
                        </td>
                      </tr>
                    )
                  })
              )}
            </tbody>
          </table>
        </div>
      ) : (
        <div className='max-h-72 overflow-y-auto bg-muted/20 dark:bg-black/50 text-foreground p-3 font-mono text-[11px] leading-relaxed selection:bg-primary/20 select-text border-t'>
          {rawBlocks.map((b, idx) => {
            if (b.type === 'header') {
              return (
                <div
                  key={idx}
                  className='text-muted-foreground/80 font-bold select-none py-1 border-b border-border/40 mt-1 first:mt-0 tracking-tight'
                >
                  {b.content}
                </div>
              )
            }
            if (b.type === 'summary' && b.summaryData) {
              const s = b.summaryData
              return (
                <div
                  key={idx}
                  className='my-1.5 rounded-md p-2 text-emerald-700 dark:text-emerald-300 shadow-2xs font-mono'
                >
                  <div className='flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold'>
                    <span>
                      sent=<span className='font-bold'>{s.sent}</span>
                    </span>
                    <span>
                      received=<span className='font-bold'>{s.received}</span>
                    </span>
                    <span className={s.lossPct > 0 ? 'text-rose-600 dark:text-rose-400 font-bold' : ''}>
                      packet-loss=<span className='font-bold'>{s.lossPct}%</span>
                    </span>
                    <span>
                      min-rtt=<span className='font-bold'>{s.minRtt}ms</span>
                    </span>
                    <span>
                      avg-rtt=<span className='font-bold'>{s.avgRtt}ms</span>
                    </span>
                    <span>
                      max-rtt=<span className='font-bold'>{s.maxRtt}ms</span>
                    </span>
                  </div>
                </div>
              )
            }
            return (
              <div
                key={idx}
                className={`py-0.5 px-1 rounded-xs transition-colors hover:bg-muted/50 ${
                  b.isTimeout ? 'text-rose-600 dark:text-rose-400 font-medium bg-rose-500/10' : 'text-foreground'
                }`}
              >
                {b.content}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
