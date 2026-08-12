import { useEffect, useRef, useState } from 'react'
import {
  CodeIcon,
  LightningBoltIcon,
  Pencil1Icon,
  TrashIcon,
  DotFilledIcon,
  ActivityLogIcon,
} from '@radix-ui/react-icons'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Device } from '@/gen/v1/device_pb'
import { deviceClient } from '@/lib/api-client'
import { useDevicesContext } from './devices-provider'

interface DeviceCardProps {
  device: Device
}

interface InterfaceItem {
  name: string
  running: boolean
  disabled: boolean
}

interface PingData {
  ms: number
  alive: boolean
}

interface TrafficData {
  rx: number
  tx: number
}

const PING_HISTORY_MAX = 40
const TRAFFIC_HISTORY_MAX = 60

function formatBps(bps: number): string {
  if (!bps || bps <= 0) return '0 bps'
  const units = ['bps', 'Kbps', 'Mbps', 'Gbps']
  let v = bps
  let i = 0
  while (v >= 1000 && i < units.length - 1) {
    v /= 1000
    i++
  }
  return `${v.toFixed(v >= 100 || i === 0 ? 0 : 1)} ${units[i]}`
}

export function DeviceCard({ device }: DeviceCardProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()

  // Live States (populated via ConnectRPC server stream)
  const [isOnline, setIsOnline] = useState<boolean>(device.enabled)
  const identityName = device.name
  const [boardName, setBoardName] = useState<string>(device.vendor)
  const [cpuUsage, setCpuUsage] = useState<number>(0)
  const [memUsage, setMemUsage] = useState<number>(0)
  const [uptime, setUptime] = useState<string>('N/A')
  const [version, setVersion] = useState<string>('N/A')
  const [pingMs, setPingMs] = useState<number>(0)
  const [pingHistory, setPingHistory] = useState<PingData[]>([])

  // Interfaces list
  const [interfaces, setInterfaces] = useState<InterfaceItem[]>([])
  const [selectedIface, setSelectedIface] = useState<string>('default')

  // Traffic
  const [rxBps, setRxBps] = useState<number>(0)
  const [txBps, setTxBps] = useState<number>(0)
  const [trafficHistory, setTrafficHistory] = useState<TrafficData[]>([])

  const pingTarget = '8.8.8.8'

  // 1. Status Stream (Metadata, Resources, Interfaces)
  useEffect(() => {
    if (!device.id || !device.enabled) return
    const controller = new AbortController()
    let active = true

    async function startStatusStream() {
      try {
        const stream = deviceClient.streamDeviceStatus(
          { id: device.id, selectedInterface: selectedIface },
          { signal: controller.signal }
        )
        for await (const frame of stream) {
          if (!active) break
          const res = frame.test
          if (res) {
            setIsOnline(res.status === 'connected' || res.status === 'online')
            if (res.boardName) setBoardName(res.boardName)
            if (res.uptime) setUptime(res.uptime)
            if (res.version) setVersion(res.version)
            if (res.cpuLoad !== undefined) setCpuUsage(res.cpuLoad)

            if (res.totalMemory && res.totalMemory > 0n) {
              const free = Number(res.freeMemory)
              const total = Number(res.totalMemory)
              const usedPct = Math.round(((total - free) / total) * 100)
              setMemUsage(usedPct)
            }

            if (res.interfaceList && res.interfaceList.length > 0) {
              const items: InterfaceItem[] = res.interfaceList.map((ifc) => ({
                id: ifc.name,
                name: ifc.name,
                type: ifc.type || (ifc.name.startsWith('ether') ? 'ether' : ifc.name.startsWith('wlan') ? 'wlan' : 'bridge'),
                disabled: Boolean(ifc.disabled),
                running: Boolean(ifc.running),
              }))
              setInterfaces(items)
              setSelectedIface((prev) => {
                if (prev === 'default') {
                  const firstEnabled = items.find((i) => !i.disabled)?.name || items[0]?.name || 'default'
                  return firstEnabled
                }
                return prev
              })
            } else if (res.interfaces && res.interfaces.length > 0) {
              const items: InterfaceItem[] = res.interfaces.map((name) => ({
                id: name,
                name: name,
                type: name.startsWith('ether') ? 'ether' : name.startsWith('wlan') ? 'wlan' : 'bridge',
                disabled: false,
                running: true,
              }))
              setInterfaces(items)
              setSelectedIface((prev) => (prev === 'default' ? res.interfaces[0] : prev))
            }
          }
        }
      } catch (err: any) {
        if (err?.name !== 'AbortError' && active) {
          setIsOnline(false)
        }
      }
    }

    startStatusStream()

    return () => {
      active = false
      controller.abort()
    }
  }, [device.id, device.enabled])

  // 2. Traffic Stream (Rx/Tx Bps per Interface)
  useEffect(() => {
    if (!device.id || !device.enabled || !selectedIface || selectedIface === 'default') return
    console.log(`[TrafficStream] Selected interface changed -> "${selectedIface}" (Device: ${device.name || device.id})`)
    const controller = new AbortController()
    let active = true

    // Reset traffic state when selected interface changes
    setRxBps(0)
    setTxBps(0)
    setTrafficHistory([])

    async function startTrafficStream() {
      try {
        console.log(`[TrafficStream] Starting stream for interface: "${selectedIface}"...`)
        const stream = deviceClient.streamInterfaceTraffic(
          { id: device.id, interfaceName: selectedIface },
          { signal: controller.signal }
        )
        for await (const frame of stream) {
          if (!active) break
          const rx = Number(frame.rxBps || 0)
          const tx = Number(frame.txBps || 0)
          console.log(`[TrafficStream] Frame received for "${selectedIface}":`, {
            rawRxBps: frame.rxBps,
            rawTxBps: frame.txBps,
            parsedRx: rx,
            parsedTx: tx,
          })
          setRxBps(rx)
          setTxBps(tx)
          setTrafficHistory((prev) => {
            const next = [...prev, { time: Date.now(), rx, tx }]
            return next.slice(-TRAFFIC_HISTORY_MAX)
          })
        }
      } catch (err: any) {
        const isAborted = controller.signal.aborted || !active || err?.name === 'AbortError' || err?.code === 1 || err?.message?.includes('aborted')
        if (!isAborted) {
          console.warn(`[TrafficStream] Stream error for "${selectedIface}":`, err)
        }
      }
    }

    startTrafficStream()

    return () => {
      console.log(`[TrafficStream] Closing stream for interface: "${selectedIface}"`)
      active = false
      controller.abort()
    }
  }, [device.id, device.enabled, selectedIface])

  // 3. Ping Stream (Latency & Target IP)
  useEffect(() => {
    if (!device.id || !device.enabled || !pingTarget) return
    const controller = new AbortController()
    let active = true

    async function startPingStream() {
      try {
        const stream = deviceClient.streamPing(
          { id: device.id, address: pingTarget },
          { signal: controller.signal }
        )
        for await (const frame of stream) {
          if (!active) break
          const latency = Number(frame.latencyMs)
          if (latency > 0) {
            setPingMs(latency)
            setPingHistory((prev) => {
              const next = [...prev, { ms: latency, alive: frame.status !== 'timeout' }]
              return next.slice(-PING_HISTORY_MAX)
            })
          }
        }
      } catch (err: any) {
        // Silent catch on abort or clean disconnect
      }
    }

    startPingStream()

    return () => {
      active = false
      controller.abort()
    }
  }, [device.id, device.enabled, pingTarget])

  // Canvas refs
  const pingCanvasRef = useRef<HTMLCanvasElement | null>(null)
  const trafficCanvasRef = useRef<HTMLCanvasElement | null>(null)

  // Render Ping Sparkline Canvas
  useEffect(() => {
    const canvas = pingCanvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.max(100, Math.round(rect.width || 240))
    const h = Math.max(20, Math.round(rect.height || 36))

    if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
      canvas.width = w * dpr
      canvas.height = h * dpr
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)

    if (pingHistory.length === 0) {
      ctx.fillStyle = '#94a3b8'
      ctx.font = '10px sans-serif'
      ctx.fillText('Testing ping...', 4, h / 2 + 3)
      return
    }

    const barW = w / PING_HISTORY_MAX
    const maxMs = Math.max(40, ...pingHistory.map((p) => p.ms))

    pingHistory.forEach((p, i) => {
      const x = w - (pingHistory.length - i) * barW
      const barH = Math.max(2, Math.min(h, (p.ms / maxMs) * h))
      const color = p.ms > 50 ? '#e8a33d' : '#2fb8ac'

      ctx.fillStyle = color
      ctx.globalAlpha = 0.35 + 0.65 * ((i + 1) / pingHistory.length)
      ctx.fillRect(x + 1, h - barH, Math.max(1, barW - 2), barH)
    })
    ctx.globalAlpha = 1
  }, [pingHistory])

  // Render Real-Time Traffic Canvas Chart
  useEffect(() => {
    const canvas = trafficCanvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.max(150, Math.round(rect.width || 320))
    const h = Math.max(60, Math.round(rect.height || 110))

    if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
      canvas.width = w * dpr
      canvas.height = h * dpr
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)

    if (trafficHistory.length < 2) {
      ctx.fillStyle = '#94a3b8'
      ctx.font = '10px sans-serif'
      ctx.fillText('Live traffic stream inactive', 8, h / 2 + 3)
      return
    }

    const pad = 4
    const actualMax = Math.max(0, ...trafficHistory.map((s) => Math.max(s.rx, s.tx)))
    const scale = Math.max(1000, actualMax)
    const stepX = (w - pad * 2) / (TRAFFIC_HISTORY_MAX - 1)
    const startI = TRAFFIC_HISTORY_MAX - trafficHistory.length

    const drawLine = (key: 'rx' | 'tx', color: string) => {
      ctx.beginPath()
      trafficHistory.forEach((s, i) => {
        const x = pad + (startI + i) * stepX
        const y = h - pad - (s[key] / scale) * (h - pad * 2)
        if (i === 0) ctx.moveTo(x, y)
        else ctx.lineTo(x, y)
      })
      ctx.strokeStyle = color
      ctx.lineWidth = 1.75
      ctx.stroke()
    }

    drawLine('rx', '#2fb8ac') // Cyan for RX
    drawLine('tx', '#e8a33d') // Amber for TX

    // Top rate label
    ctx.fillStyle = '#64748b'
    ctx.font = '10px ui-monospace, monospace'
    ctx.fillText(`${formatBps(actualMax)} peak`, pad + 2, 12)
  }, [trafficHistory])

  return (
    <article className='flex flex-col rounded-xl border bg-card p-4 text-card-foreground shadow-sm transition-all hover:shadow-md'>
      {/* ===== Card Header ===== */}
      <header className='flex items-start justify-between gap-2 border-b pb-3'>
        <div className='flex items-center gap-2 min-w-0'>
          <span
            className={`inline-flex items-center justify-center rounded-full p-0.5 ${
              isOnline ? 'text-emerald-500 animate-pulse' : 'text-slate-400'
            }`}
            title={isOnline ? 'Online' : 'Offline'}
          >
            <DotFilledIcon className='h-5 w-5' />
          </span>

          <div className='min-w-0'>
            <h3 className='font-semibold text-base leading-none truncate' title={identityName}>
              {identityName}
            </h3>
            <p className='text-xs text-muted-foreground font-mono mt-1 truncate'>
              {device.host}:{device.port} {device.sshPort ? `(SSH ${device.sshPort})` : ''}
            </p>
          </div>
        </div>

        <div className='flex items-center gap-1 shrink-0'>
          {boardName && (
            <Badge variant='outline' className='text-[10px] font-normal uppercase hidden sm:inline-flex'>
              {boardName}
            </Badge>
          )}

          {/* Action Buttons using Radix UI Icons */}
          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7 text-muted-foreground hover:text-emerald-600'
            title='Terminal SSH'
            onClick={() => {
              setCurrentRow(device)
              setOpen('terminal')
            }}
          >
            <CodeIcon className='h-4 w-4' />
          </Button>

          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7 text-muted-foreground hover:text-amber-500'
            title='Test Connection'
            onClick={() => {
              setCurrentRow(device)
              setOpen('test')
            }}
          >
            <LightningBoltIcon className='h-4 w-4' />
          </Button>

          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7 text-muted-foreground hover:text-primary'
            title='Edit Device'
            onClick={() => {
              setCurrentRow(device)
              setOpen('edit')
            }}
          >
            <Pencil1Icon className='h-4 w-4' />
          </Button>

          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7 text-muted-foreground hover:text-destructive'
            title='Delete Device'
            onClick={() => {
              setCurrentRow(device)
              setOpen('delete')
            }}
          >
            <TrashIcon className='h-4 w-4' />
          </Button>
        </div>
      </header>

      {/* ===== Metrics Row (CPU & RAM) ===== */}
      <section className='grid grid-cols-2 gap-3 py-3 border-b text-xs'>
        <div>
          <div className='flex justify-between mb-1 text-muted-foreground'>
            <span>CPU</span>
            <span className='font-mono font-medium text-foreground'>{cpuUsage}%</span>
          </div>
          <div className='h-1.5 w-full bg-secondary rounded-full overflow-hidden'>
            <div
              className={`h-full transition-all duration-500 ${
                cpuUsage >= 90
                  ? 'bg-rose-500'
                  : cpuUsage >= 70
                  ? 'bg-amber-500'
                  : 'bg-emerald-500'
              }`}
              style={{ width: `${Math.min(100, cpuUsage)}%` }}
            />
          </div>
        </div>

        <div>
          <div className='flex justify-between mb-1 text-muted-foreground'>
            <span>Memory</span>
            <span className='font-mono font-medium text-foreground'>{memUsage}%</span>
          </div>
          <div className='h-1.5 w-full bg-secondary rounded-full overflow-hidden'>
            <div
              className={`h-full transition-all duration-500 ${
                memUsage >= 90
                  ? 'bg-rose-500'
                  : memUsage >= 70
                  ? 'bg-amber-500'
                  : 'bg-emerald-500'
              }`}
              style={{ width: `${Math.min(100, memUsage)}%` }}
            />
          </div>
        </div>

        <div className='flex justify-between col-span-1 text-[11px] text-muted-foreground pt-1'>
          <span>Uptime:</span>
          <span className='font-mono text-foreground'>{uptime}</span>
        </div>

        <div className='flex justify-between col-span-1 text-[11px] text-muted-foreground pt-1'>
          <span>RouterOS:</span>
          <span className='font-mono text-foreground'>{version}</span>
        </div>
      </section>

      {/* ===== Ping Readout & Sparkline Canvas ===== */}
      <section className='py-2.5 border-b flex items-center justify-between gap-2'>
        <div className='flex items-baseline gap-1 text-xs'>
          <ActivityLogIcon className='h-3.5 w-3.5 text-muted-foreground' />
          <span className='font-mono font-bold text-sm text-foreground'>{pingMs}</span>
          <span className='text-[10px] text-muted-foreground'>ms</span>
          <span className='text-[10px] text-muted-foreground ml-1 hidden sm:inline'>→ 8.8.8.8</span>
        </div>
        <div className='w-32 h-7 bg-muted/30 rounded overflow-hidden'>
          <canvas ref={pingCanvasRef} className='w-full h-full block' />
        </div>
      </section>

      {/* ===== Interface Dropdown Section ===== */}
      <section className='py-2.5 border-b space-y-1.5'>
        <div className='flex items-center justify-between text-xs'>
          <span className='font-medium text-muted-foreground'>Ethernet Interface</span>
          <span className='text-[10px] text-muted-foreground'>Select to monitor</span>
        </div>
        <Select value={selectedIface} onValueChange={(val) => {
          console.log(`[UI Select] User selected interface from dropdown: "${val}"`)
          setSelectedIface(val)
        }}>
          <SelectTrigger className='h-8 text-xs font-mono'>
            <SelectValue placeholder='Select interface...' />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectLabel className='text-[11px] text-muted-foreground'>Enabled Interfaces</SelectLabel>
              {interfaces
                .filter((i) => !i.disabled)
                .map((iface) => (
                  <SelectItem key={iface.name} value={iface.name} className='text-xs font-mono'>
                    <span className='flex items-center gap-1.5'>
                      <span className={`h-1.5 w-1.5 rounded-full ${iface.running ? 'bg-emerald-500' : 'bg-slate-400'}`} />
                      {iface.name}
                    </span>
                  </SelectItem>
                ))}
            </SelectGroup>
            <SelectGroup>
              <SelectLabel className='text-[11px] text-muted-foreground'>Disabled Interfaces</SelectLabel>
              {interfaces
                .filter((i) => i.disabled)
                .map((iface) => (
                  <SelectItem key={iface.name} value={iface.name} className='text-xs font-mono opacity-50'>
                    {iface.name} (disabled)
                  </SelectItem>
                ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </section>

      {/* ===== Traffic Monitor Section ===== */}
      <section className='pt-2.5 space-y-2'>
        <div className='flex items-center justify-between text-xs'>
          <span className='font-medium flex items-center gap-1.5'>
            Traffic <span className='font-mono text-emerald-600 dark:text-emerald-400'>{selectedIface}</span>
            {(() => {
              const ifc = interfaces.find((i) => i.name === selectedIface)
              if (!ifc) return null
              if (ifc.disabled) {
                return <span className='text-[10px] px-1.5 py-0.5 rounded bg-rose-500/10 text-rose-500 font-medium'>(disabled)</span>
              }
              if (!ifc.running) {
                return <span className='text-[10px] px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-500 font-medium'>(link down)</span>
              }
              return null
            })()}
          </span>
          <div className='flex items-center gap-2 text-[11px] font-mono'>
            <span className='text-emerald-600 dark:text-emerald-400 font-medium'>
              RX <b>{formatBps(rxBps)}</b>
            </span>
            <span className='text-amber-600 dark:text-amber-400 font-medium'>
              TX <b>{formatBps(txBps)}</b>
            </span>
          </div>
        </div>

        <div className='w-full h-24 bg-muted/40 rounded-lg p-1 overflow-hidden border border-border/50'>
          <canvas ref={trafficCanvasRef} className='w-full h-full block' />
        </div>
      </section>
    </article>
  )
}
