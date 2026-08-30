import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { DeviceInterfaceItem, TrafficDataPoint } from '../../types'
import { formatBps } from '../../lib/formatters'
import { DeviceTrafficChart } from './device-traffic-chart'
import { ArrowDown, ArrowUp } from 'lucide-react'

interface DeviceCardTrafficProps {
  interfaces: DeviceInterfaceItem[]
  selectedIface: string
  setSelectedIface: (name: string) => void
  rxBps: number
  txBps: number
  trafficHistory: TrafficDataPoint[]
}

export function DeviceCardTraffic({
  interfaces,
  selectedIface,
  setSelectedIface,
  rxBps,
  txBps,
  trafficHistory,
}: DeviceCardTrafficProps) {
  const currentIface = interfaces.find((i) => i.name === selectedIface)

  return (
    <section className='pt-2.5 space-y-2'>
      {/* Interface Selector & Live Rate Badges */}
      <div className='flex items-center justify-between gap-2 text-xs'>
        <div className='flex items-center gap-1.5 min-w-0'>
          <span className='font-medium text-muted-foreground shrink-0 text-[11px]'>
            Traffic
          </span>

          <Select
            value={selectedIface}
            onValueChange={(val) => {
              setSelectedIface(val)
            }}
          >
            <SelectTrigger className='h-7 px-2 text-xs font-mono w-auto min-w-[90px] max-w-[140px] bg-background/80'>
              <SelectValue placeholder='Select interface...' />
            </SelectTrigger>
            <SelectContent>
              {(() => {
                const enabled = interfaces.filter((i) => !i.disabled)
                const disabled = interfaces.filter((i) => i.disabled)

                return (
                  <>
                    {enabled.length > 0 && (
                      <SelectGroup>
                        {disabled.length > 0 && (
                          <SelectLabel className='text-[10px] text-muted-foreground uppercase'>
                            Enabled Interfaces
                          </SelectLabel>
                        )}
                        {enabled.map((iface) => (
                          <SelectItem
                            key={iface.name}
                            value={iface.name}
                            className='text-xs font-mono'
                          >
                            <span className='flex items-center gap-1.5'>
                              <span
                                className={`h-1.5 w-1.5 rounded-full ${
                                  iface.running ? 'bg-emerald-500' : 'bg-slate-400'
                                }`}
                              />
                              {iface.name}
                              {!iface.running && (
                                <span className='text-[10px] text-muted-foreground ml-1'>
                                  (down)
                                </span>
                              )}
                            </span>
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    )}

                    {disabled.length > 0 && (
                      <SelectGroup>
                        <SelectLabel className='text-[10px] text-muted-foreground uppercase'>
                          Disabled Interfaces
                        </SelectLabel>
                        {disabled.map((iface) => (
                          <SelectItem
                            key={iface.name}
                            value={iface.name}
                            className='text-xs font-mono opacity-60'
                          >
                            <span className='flex items-center gap-1.5'>
                              <span className='h-1.5 w-1.5 rounded-full bg-rose-400' />
                              {iface.name}
                              <span className='text-[10px] text-rose-500 ml-1'>
                                (disabled)
                              </span>
                            </span>
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    )}
                  </>
                )
              })()}
            </SelectContent>
          </Select>

          {currentIface?.disabled && (
            <span className='text-[10px] px-1.5 py-0.5 rounded bg-rose-500/10 text-rose-500 font-medium'>
              (disabled)
            </span>
          )}
          {currentIface && !currentIface.disabled && !currentIface.running && (
            <span className='text-[10px] px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-500 font-medium'>
              (down)
            </span>
          )}
        </div>

        {/* Live Throughput RX / TX */}
        <div className='flex items-center gap-2 text-[11px] font-mono shrink-0'>
          <span
            className='inline-flex items-center gap-0.5 text-emerald-600 dark:text-emerald-400 font-medium'
            title='Download Throughput (RX)'
          >
            <ArrowDown className='h-3 w-3' />
            <b>{formatBps(rxBps)}</b>
          </span>
          <span
            className='inline-flex items-center gap-0.5 text-amber-600 dark:text-amber-400 font-medium'
            title='Upload Throughput (TX)'
          >
            <ArrowUp className='h-3 w-3' />
            <b>{formatBps(txBps)}</b>
          </span>
        </div>
      </div>

      {/* Traffic Canvas Chart Container */}
      <div className='w-full h-24 bg-muted/30 rounded-lg p-1 overflow-hidden border border-border/40'>
        <DeviceTrafficChart trafficHistory={trafficHistory} />
      </div>
    </section>
  )
}
